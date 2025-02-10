from abc import ABC, abstractmethod
from dataclasses import dataclass
from threading import Lock
from typing import Any, Dict, List, Optional

from platform_common.utils import save_dict
from pydantic import BaseModel


class LLMBase(ABC):
    def __init__(self, model_name: str, track_usage_at: Optional[str] = None):
        self.usage = dict()
        self.lock = Lock()
        self.model_name = model_name
        self.track_usage_at = track_usage_at

    @abstractmethod
    def completion(
        self, prompt: str, system_prompt: Optional[str] = None, **kwargs
    ) -> str:
        pass

    def verify_access(self):
        self.client.models.list()

    def track_usage(self, model_name: str, current_usage):
        with self.lock:
            if model_name not in self.usage:
                self.usage[model_name] = current_usage
            else:
                self.usage[model_name] = self.usage[model_name] + current_usage

            if self.track_usage_at:
                save_dict(self.usage, self.track_usage_at.__dict__)

    def _process_prompt(
        self,
        prompt: str,
        system_prompt: Optional[str] = None,
        metadata: Optional[dict] = None,
        **completion_kwargs,
    ):
        response, usage = self.client.completion(
            prompt=prompt, system_prompt=system_prompt, **completion_kwargs
        )
        return response, usage, metadata

    def run_and_collect_results(
        self, tasks_prompt: List[Dict[str, Any]], parallelize: bool = True
    ):
        """
        Function to process the prompts parallely
        args: task_prompt: List of prompts to process, List[Dict[str, str]]
            Format of each argument:
                prompt: Generation prompt to the LLM.
                system_prompt: system prompt to the LLM.
                completion_kwargs: passed to the completion function
        """
        data_points = []
        if parallelize:
            from concurrent.futures import ThreadPoolExecutor, as_completed

            with ThreadPoolExecutor() as executor:
                futures = []

                # Submit arguments to the executor
                for task in tasks_prompt:
                    future = executor.submit(
                        self._process_prompt,
                        task["prompt"],
                        task.get("system_prompt"),
                        task.get("metadata", None),
                        **(task.get("completion_kwargs", {})),
                    )
                    futures.append(future)

                # Wait for all arguments to complete and handle exceptions
                for future in as_completed(futures):
                    try:
                        response, _, metadata = future.result()
                        data_points.append((response, metadata))

                    except Exception as e:
                        # TODO(gautam): Add logging of error somewhere
                        pass

        else:
            for task in tasks_prompt:
                try:
                    response, _, metadata = self._process_prompt(
                        task["prompt"],
                        task.get("system_prompt"),
                        task.get("metadata", None),
                        **(task.get("kwargs") or {}),
                    )
                    data_points.append((response, metadata))

                except Exception as e:
                    # TODO(gautam): Add logging of error somewhere
                    pass

        return data_points


class OpenAILLM(LLMBase):
    @dataclass
    class Usage:
        completion_tokens: int
        prompt_tokens: int
        total_tokens: int

        def __add__(self, other):
            if not isinstance(other, OpenAILLM.Usage):
                return NotImplementedError("Unsupported operation")
            return OpenAILLM.Usage(
                completion_tokens=self.completion_tokens + other.completion_tokens,
                prompt_tokens=self.prompt_tokens + other.prompt_tokens,
                total_tokens=self.total_tokens + other.total_tokens,
            )

    def __init__(
        self,
        api_key: str,
        model_name: str = "gpt-4o",
        base_url: Optional[str] = None,
        track_usage_at: Optional[str] = None,
    ):
        from openai import OpenAI

        super().__init__(model_name, track_usage_at)
        self.client = OpenAI(base_url=base_url, api_key=api_key)
        self.verify_access()

    def completion(
        self,
        prompt,
        system_prompt=None,
        temperature: float = 0.8,
        response_format: Optional[BaseModel] = None,
        **kwargs,
    ):
        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})

        messages.append({"role": "user", "content": prompt})

        if response_format:
            completion = self.client.beta.chat.completions.parse(
                model=self.model_name,
                messages=messages,
                temperature=temperature,
                response_format=response_format,
            )
            res = completion.choices[0].message.parsed
        else:
            completion = self.client.chat.completions.create(
                model=self.model_name,
                messages=messages,
                temperature=temperature,
            )
            res = completion.choices[0].message.content

        current_usage = completion.usage
        self.track_usage(
            self.model_name,
            self.Usage(
                completion_tokens=current_usage.completion_tokens,
                prompt_tokens=current_usage.prompt_tokens,
                total_tokens=current_usage.total_tokens,
            ),
        )
        return res, current_usage


class CohereLLM(LLMBase):
    @dataclass
    class Usage:
        input_tokens: int
        output_tokens: int

        def __add__(self, other):
            if not isinstance(other, CohereLLM.Usage):
                return NotImplementedError("Unsupported operation")
            return CohereLLM.Usage(
                input_tokens=self.input_tokens + other.input_tokens,
                output_tokens=self.output_tokens + other.output_tokens,
            )

    def __init__(
        self,
        api_key: str,
        model_name: str = "command-r-plus",
        base_url: Optional[str] = None,
        track_usage_at: Optional[str] = None,
    ):
        from cohere import ClientV2

        super().__init__(model_name, track_usage_at)
        self.client = ClientV2(api_key=api_key, base_url=base_url)
        self.verify_access()

    def completion(self, prompt, system_prompt: Optional[str] = None, **kwargs):

        messages = []
        if system_prompt:
            messages.append({"role": "system", "content": system_prompt})

        messages.append({"role": "user", "content": prompt})

        completion = self.client.chat(model=self.model_name, messages=messages)

        response_text = completion.message.content[0].text
        current_usage = completion.usage.billed_units
        self.track_usage(
            self.model_name,
            self.Usage(
                input_tokens=current_usage.input_tokens,
                output_tokens=current_usage.output_tokens,
            ),
        )
        return response_text, current_usage


llm_classes = {
    "openai": OpenAILLM,
    "cohere": CohereLLM,
}
