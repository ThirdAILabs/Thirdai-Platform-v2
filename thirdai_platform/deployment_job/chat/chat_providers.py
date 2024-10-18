from urllib.parse import urljoin

from deployment_job.chat.chat_interface import ChatInterface
from langchain_openai import ChatOpenAI  # type: ignore
from thirdai import neural_db as ndb


class OpenAIChat(ChatInterface):
    def __init__(
        self,
        db: ndb.NeuralDB,
        chat_history_sql_uri: str,
        key: str,
        top_k: int = 5,
        model: str = "gpt-4o-mini",
        temperature: float = 0.2,
        chat_prompt: str = "Answer the user's questions based on the below context:",
        query_reformulation_prompt: str = "Given the above conversation, generate a search query that would help retrieve relevant sources for responding to the last message.",
        **kwargs,
    ):
        # Set instance variables necessary for self.llm() before calling super().__init__(),
        # because super().__init__() calls self.llm()
        self.model = model
        self.key = key
        self.temperature = temperature

        super().__init__(
            db, chat_history_sql_uri, top_k, chat_prompt, query_reformulation_prompt
        )

    def llm(self):
        return ChatOpenAI(
            model=self.model,
            temperature=self.temperature,
            openai_api_key=self.key,
        )


class OnPremChat(ChatInterface):
    def __init__(
        self,
        db: ndb.NeuralDB,
        chat_history_sql_uri: str,
        base_url: str,
        key: str = None,
        top_k: int = 5,
        chat_prompt: str = "Answer the user's questions based on the below context:",
        query_reformulation_prompt: str = "Given the above conversation, generate a search query that would help retrieve relevant sources for responding to the last message.",
        **kwargs,
    ):
        # Set instance variables necessary for self.llm() before calling super().__init__(),
        # because super().__init__() calls self.llm()
        self.base_url = base_url
        self.key = key

        super().__init__(
            db, chat_history_sql_uri, top_k, chat_prompt, query_reformulation_prompt
        )

    def llm(self):
        headers = {"Authorization": f"Bearer {self.key}"} if self.key else {}
        return ChatOpenAI(
            base_url=urljoin(self.base_url, "on-prem-llm"),
            default_headers=headers,
        )

    def set_token(self, token: str):
        self.token = token
        self.conversational_retrieval_chain = (
            self._create_conversational_retrieval_chain()
        )

    def chat(self, user_input: str, session_id: str, token: str = None, **kwargs):
        if token:
            self.set_token(token)
        return super().chat(user_input, session_id, **kwargs)

    async def stream_chat(
        self, user_input: str, session_id: str, token: str = None, **kwargs
    ):
        if token:
            self.set_token(token)
        async for chunk in super().stream_chat(user_input, session_id, **kwargs):
            yield chunk
