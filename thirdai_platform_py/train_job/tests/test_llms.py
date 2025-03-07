import unittest
from unittest.mock import MagicMock, patch

import pytest

from ..llm.api_clients import CohereLLM, MockLLM, OpenAILLM, TokenCount


class TestLLMBase(unittest.TestCase):
    def setUp(self):
        self.test_prompt = "Test prompt"
        self.test_system_prompt = "Test system prompt"
        self.test_api_key = "test_key"


class TestMockLLM(TestLLMBase):
    def setUp(self):
        super().setUp()
        self.llm = MockLLM(api_key=self.test_api_key)

    def test_completion(self):
        response, token_count = self.llm.completion(self.test_prompt)
        self.assertEqual(
            response, "Mocked response from the llm to test the generation flow"
        )
        self.assertEqual(token_count, TokenCount(0, 0, 0))


class TestOpenAILLM(TestLLMBase):
    def setUp(self):
        super().setUp()
        self.patcher = patch("openai.OpenAI")
        self.mock_openai = self.patcher.start()
        self.llm = OpenAILLM(api_key=self.test_api_key)

    def tearDown(self):
        self.patcher.stop()

    def test_completion_success(self):
        # Mocking the OpenAI completion response
        mock_completion = MagicMock()
        mock_completion.choices = [
            MagicMock(message=MagicMock(content="Test response"))
        ]
        mock_completion.usage = MagicMock(
            completion_tokens=10, prompt_tokens=5, total_tokens=15
        )
        self.llm.client.chat.completions.create.return_value = mock_completion

        response, token_count = self.llm.completion(self.test_prompt)

        self.assertEqual(response, "Test response")
        self.assertEqual(token_count.completion_tokens, 10)
        self.assertEqual(token_count.prompt_tokens, 5)
        self.assertEqual(token_count.total_tokens, 15)

    def test_completion_with_system_prompt(self):
        mock_completion = MagicMock()
        mock_completion.choices = [
            MagicMock(message=MagicMock(content="Test response"))
        ]
        mock_completion.usage = MagicMock(
            completion_tokens=1, prompt_tokens=2, total_tokens=3
        )
        self.llm.client.chat.completions.create.return_value = mock_completion

        response, _ = self.llm.completion(
            self.test_prompt, system_prompt=self.test_system_prompt
        )

        # verifying the system prompt inclusion in the messages.
        called_args = self.llm.client.chat.completions.create.call_args[1]
        messages = called_args["messages"]
        self.assertEqual(len(messages), 2)
        self.assertEqual(messages[0]["role"], "system")
        self.assertEqual(messages[0]["content"], self.test_system_prompt)


class TestCohereLLM(TestLLMBase):
    def setUp(self):
        super().setUp()
        self.patcher = patch("cohere.ClientV2")
        self.mock_cohere = self.patcher.start()
        self.llm = CohereLLM(api_key=self.test_api_key)

    def tearDown(self):
        self.patcher.stop()

    def test_completion_success(self):
        # Mocking the Cohere completion response
        mock_completion = MagicMock()
        mock_completion.message.content = [MagicMock(text="Test response")]
        mock_completion.usage.billed_units = MagicMock(input_tokens=5, output_tokens=10)
        self.llm.client.chat.return_value = mock_completion

        response, token_count = self.llm.completion(self.test_prompt)

        self.assertEqual(response, "Test response")
        self.assertEqual(token_count.prompt_tokens, 5)
        self.assertEqual(token_count.completion_tokens, 10)
        self.assertEqual(token_count.total_tokens, 15)

    def test_completion_with_system_prompt(self):
        mock_completion = MagicMock()
        mock_completion.message.content = [MagicMock(text="Test response")]
        mock_completion.usage.billed_units = MagicMock(input_tokens=5, output_tokens=10)
        self.llm.client.chat.return_value = mock_completion

        response, _ = self.llm.completion(
            self.test_prompt, system_prompt=self.test_system_prompt
        )

        # verifying the system prompt inclusion in the messages.
        called_args = self.llm.client.chat.call_args[1]
        messages = called_args["messages"]
        self.assertEqual(len(messages), 2)
        self.assertEqual(messages[0]["role"], "system")
        self.assertEqual(messages[0]["content"], self.test_system_prompt)


@pytest.mark.parametrize(
    "llm_class,api_key", [(OpenAILLM, "test_key"), (CohereLLM, "test_key")]
)
def test_llm_initialization(llm_class, api_key):
    """Test that all LLM classes can be initialized"""
    with patch(
        "openai.OpenAI" if llm_class == OpenAILLM else "cohere.ClientV2"
    ) as mock_client:
        # Mocking the models.list() call to prevent the api-key verification
        mock_client.return_value.models = MagicMock()
        mock_client.return_value.models.list = MagicMock()

        with patch.object(llm_class, "verify_access"):
            llm = llm_class(api_key=api_key)
            assert llm.model_name is not None

            # Verify that client was initialized with correct api_key
            mock_client.assert_called_once_with(base_url=None, api_key=api_key)
