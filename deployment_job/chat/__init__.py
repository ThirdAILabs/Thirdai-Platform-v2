import os

from chat.openai import OpenAIChat

llm_providers = {"openai": OpenAIChat}

llm_default_keys = {"openai": os.getenv("OPENAI", None)}
