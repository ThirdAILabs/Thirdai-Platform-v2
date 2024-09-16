from abc import ABC, abstractmethod
from typing import List, Union

from chat.ndbv2_vectorstore import NeuralDBV2VectorStore
from langchain.chains.combine_documents import create_stuff_documents_chain
from langchain.docstore.document import Document
from langchain.vectorstores import NeuralDBVectorStore
from langchain_community.chat_message_histories import SQLChatMessageHistory
from langchain_core.language_models.llms import LLM
from langchain_core.messages import AIMessage
from langchain_core.output_parsers import StrOutputParser
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.runnables import RunnableBranch, RunnablePassthrough
from thirdai import neural_db as ndb
from thirdai import neural_db_v2 as ndbv2

import logging

logging.basicConfig(level=logging.INFO)


class ChatInterface(ABC):
    def __init__(
        self,
        db: Union[ndb.NeuralDB, ndbv2.NeuralDB],
        chat_history_sql_uri: str,
        top_k: int = 5,
        chat_prompt: str = "Answer the user's questions based on the below context:",
        query_reformulation_prompt: str = "Given the above conversation, generate a search query that would help retrieve relevant sources for responding to the last message.",
        **kwargs,
    ):
        self.chat_history_sql_uri = chat_history_sql_uri
        if isinstance(db, ndb.NeuralDB):
            vectorstore = NeuralDBVectorStore(db)
        elif isinstance(db, ndbv2.NeuralDB):
            vectorstore = NeuralDBV2VectorStore(db)
        else:
            raise ValueError(f"Cannot support db of type {type(db)}")

        retriever = vectorstore.as_retriever(search_kwargs={"k": top_k})

        query_transform_prompt = ChatPromptTemplate.from_messages(
            [
                MessagesPlaceholder(variable_name="messages"),
                (
                    "user",
                    query_reformulation_prompt,
                ),
            ]
        )

        query_transforming_retriever_chain = RunnableBranch(
            (
                lambda x: len(x.get("messages", [])) == 1,
                # If only one message, then we just pass that message's content to retriever
                (lambda x: x["messages"][-1].content) | retriever,
            ),
            # If messages, then we pass inputs to LLM chain to transform the query, then pass to retriever
            query_transform_prompt | self.llm() | StrOutputParser() | retriever,
        ).with_config(run_name="chat_retriever_chain")

        question_answering_prompt = ChatPromptTemplate.from_messages(
            [
                (
                    "system",
                    chat_prompt + "\n\n{context}",
                ),
                MessagesPlaceholder(variable_name="messages"),
            ]
        )

        document_chain = create_stuff_documents_chain(
            self.llm(), question_answering_prompt
        )

        self.conversational_retrieval_chain = RunnablePassthrough.assign(
            context=query_transforming_retriever_chain
            | ChatInterface.parse_retriever_output,
        ).assign(
            answer=document_chain,
        )

    @abstractmethod
    def llm(self) -> LLM:
        raise NotImplementedError()

    def parse_retriever_output(self, documents: List[Document]):
        top_k_docs = documents
        filtered_docs = []

        for doc in top_k_docs:
            metadata = doc.metadata.get('metadata', {})

            doc_id = doc.metadata.get('id', '')
            file_path = doc.metadata.get('source', '')
            page = metadata.get('page', 1)  # Default to 1 if page not specified

            filtered_doc_info = {
                'reference_type': 'File',
                'id': doc_id,
                'file_path': file_path,
                'page': page
            }

            filtered_docs.append(filtered_doc_info)

        # Clear metadata to save memory
        for doc in top_k_docs:
            doc.metadata = {}

        # Log the extracted references
        logging.info(
            f"Extracted References: {filtered_docs}"
        )

        # Optionally, store the references if needed later
        self.references = filtered_docs

        return top_k_docs

    def get_chat_history(self, session_id: str, **kwargs):
        chat_history = SQLChatMessageHistory(
            session_id=session_id, connection_string=self.chat_history_sql_uri
        )
        chat_history_list = [
            {
                "content": message.content,
                "sender": "AI" if isinstance(message, AIMessage) else "human",
            }
            for message in chat_history.messages
        ]
        return chat_history_list

    def chat(self, user_input: str, session_id: str, **kwargs):
        chat_history = SQLChatMessageHistory(
            session_id=session_id, connection_string=self.chat_history_sql_uri
        )
        chat_history.add_user_message(user_input)

        # Invoke the conversational retrieval chain
        response = self.conversational_retrieval_chain.invoke(
            {"messages": chat_history.messages}
        )

        # Add AI's response to the chat history
        chat_history.add_ai_message(response["answer"])

        # Log the references
        logging.info(
            f"References for this chat: {self.references}"
        )

        return response["answer"]
