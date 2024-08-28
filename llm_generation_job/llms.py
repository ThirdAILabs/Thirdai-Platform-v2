import aiohttp
import json
from typing import AsyncGenerator, Dict, Any


class DynamicLLM:
    def __init__(self, config: Dict[str, Any]):
        self.config = config

    async def stream(self, key: str, query: str, model: str) -> AsyncGenerator[str, None]:
        url = self.config["api_url"]
        headers = {
            header_name: header_value.format(key=key)
            for header_name, header_value in self.config["headers"].items()
        }
        body = {
            key: (value.format(model=model, query=query) if isinstance(value, str) else value)
            for key, value in self.config["body_template"].items()
        }

        async with aiohttp.ClientSession() as session:
            async with session.post(url, headers=headers, json=body) as response:
                if response.status == 200:
                    async for chunk in response.content:
                        chunk_str = chunk.decode("utf8").strip()
                        try:
                            async for content in self._dynamic_parse_output(chunk_str):
                                yield content
                        except Exception as e:
                            raise Exception(f"Error processing response chunk: {e}")
                else:
                    error_message = await response.text()
                    raise Exception(f"API request failed: {error_message}")

    async def _dynamic_parse_output(self, chunk: str) -> AsyncGenerator[str, None]:
        parsing_rules = self.config["parsing_rules"]
        data_chunks = self._split_and_strip(chunk, parsing_rules)

        for data_chunk in data_chunks:
            try:
                data = json.loads(data_chunk)
            except json.JSONDecodeError:
                raise Exception(f"Invalid JSON format: {data_chunk}")

            if not self._check_conditions(data, parsing_rules.get("condition", {})):
                continue

            content = self._extract_content(data, parsing_rules)
            if content:
                yield content

            if self._check_termination(data, parsing_rules.get("termination_check", {})):
                return

    def _split_and_strip(self, chunk: str, parsing_rules: Dict[str, Any]) -> list[str]:
        prefix = parsing_rules.get("prefix", "")
        delimiter = parsing_rules.get("delimiter", "\n")
        
        if chunk.startswith(prefix):
            chunk = chunk[len(prefix):]

        return [data.strip() for data in chunk.split(delimiter) if data.strip()]

    def _extract_content(self, data: Dict[str, Any], parsing_rules: Dict[str, Any]) -> str:
        content_key_path = parsing_rules.get("content_key", [])
        for key in content_key_path:
            if isinstance(data, list) and isinstance(key, int) and key < len(data):
                data = data[key]
            elif isinstance(data, dict) and key in data:
                data = data[key]
            else:
                return ""

        return data if isinstance(data, str) else ""

    def _check_conditions(self, data: Dict[str, Any], condition: Dict[str, Any]) -> bool:
        key, value = condition.get("key"), condition.get("value")
        return data.get(key) == value if key else True

    def _check_termination(self, data: Dict[str, Any], termination_check: Dict[str, Any]) -> bool:
        key, expected_value = termination_check.get("key"), termination_check.get("value")
        return data.get(key) == expected_value if key else False
