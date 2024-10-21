from abc import ABC, abstractmethod
from typing import AsyncGenerator, Callable, List, Union

from deployment_job.chat.ndbv2_vectorstore import NeuralDBV2VectorStore
from langchain.chains.combine_documents import create_stuff_documents_chain
from langchain.docstore.document import Document
from langchain.vectorstores import NeuralDBVectorStore
from langchain_community.chat_message_histories import SQLChatMessageHistory
from langchain_core.language_models.llms import LLM
from langchain_core.messages import AIMessage
from langchain_core.output_parsers import StrOutputParser
from langchain_core.prompts import ChatPromptTemplate, MessagesPlaceholder
from langchain_core.runnables import RunnableBranch, RunnablePassthrough

pass
from thirdai import neural_db as ndb
from thirdai import neural_db_v2 as ndbv2


class ChatInterface(ABC):
    def __init__(
        self,
        db: Union[ndb.NeuralDB, ndbv2.NeuralDB],
        chat_history_sql_uri: str,
        top_k: int = 5,
        chat_prompt: str = "Answer the user's questions based on the below context:",
        query_reformulation_prompt: str = "Given the above conversation, generate a search query that would help retrieve relevant sources for responding to the last message.",
    ):
        self.chat_history_sql_uri = chat_history_sql_uri
        self.top_k = top_k
        self.chat_prompt = chat_prompt
        self.query_reformulation_prompt = query_reformulation_prompt

        if isinstance(db, ndb.NeuralDB):
            self.vectorstore = NeuralDBVectorStore(db)
        elif isinstance(db, ndbv2.NeuralDB):
            self.vectorstore = NeuralDBV2VectorStore(db)
        else:
            raise ValueError(f"Cannot support db of type {type(db)}")

    def create_chain(
        self,
        llm: Callable[[], LLM],
    ):
        retriever = self.vectorstore.as_retriever(search_kwargs={"k": self.top_k})

        query_transform_prompt = ChatPromptTemplate.from_messages(
            [
                MessagesPlaceholder(variable_name="messages"),
                (
                    "user",
                    self.query_reformulation_prompt,
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
            query_transform_prompt | llm() | StrOutputParser() | retriever,
        ).with_config(run_name="chat_retriever_chain")

        question_answering_prompt = ChatPromptTemplate.from_messages(
            [
                (
                    "system",
                    self.chat_prompt + "\n\n{context}",
                ),
                MessagesPlaceholder(variable_name="messages"),
            ]
        )

        document_chain = create_stuff_documents_chain(llm(), question_answering_prompt)

        return RunnablePassthrough.assign(
            context=query_transforming_retriever_chain
            | ChatInterface.parse_retriever_output,
        ).assign(
            answer=document_chain,
        )

    @staticmethod
    def parse_retriever_output(documents: List[Document]):
        top_k_docs = documents

        # The chatbot currently doesn't utilize any metadata, so we delete it to save memory
        for doc in top_k_docs:
            doc.metadata = {}

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

    @abstractmethod
    async def stream_chat(
        self, user_input: str, session_id: str, access_token: str = None, **kwargs
    ):
        raise NotImplementedError()

    async def stream_chat_helper(
        self,
        user_input: str,
        session_id: str,
        llm: Callable[[], LLM],
    ) -> AsyncGenerator[str, None]:
        chain = self.create_chain(llm)
        chat_history = SQLChatMessageHistory(
            session_id=session_id, connection_string=self.chat_history_sql_uri
        )
        chat_history.add_user_message(user_input)

        response_chunks = []
        async for chunk in chain.astream({"messages": chat_history.messages}):
            if "answer" in chunk:
                response_chunks.append(chunk["answer"])
                yield chunk["answer"]

        full_response = "".join(response_chunks)
        chat_history.add_ai_message(full_response)
