from deployment_job.chat.chat_providers import OnPremChat, OpenAIChat

llm_providers = {"openai": OpenAIChat, "on-prem": OnPremChat}
