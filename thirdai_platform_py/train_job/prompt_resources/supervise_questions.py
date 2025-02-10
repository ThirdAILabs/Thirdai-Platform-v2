from typing import List

from pydantic import BaseModel

user_prompt = """
** Generate {questions_per_chunk} questions based on the following context.**
- Avoid simple keyword-based questions; instead, focus on deeper understanding, implications, or connections within the text.
- Aim for a mix of factual, analytical, and open-ended questions.

context:
```{chunk_text}```
"""

system_prompt = """** You are an AI model designed to generate insightful questions from a given text.**"""


class Response(BaseModel):
    questions: List[str]
