from typing import List

from platform_common.pydantic_models.llm_config import LLMProvider
from pydantic import BaseModel

user_prompt = {
    LLMProvider.openai: """
** Generate {questions_per_chunk} questions based on the following context.**
- Avoid simple keyword-based questions; instead, focus on deeper understanding, implications, or connections within the text.
- Aim for a mix of factual, analytical, and open-ended questions.

context:
```{chunk_text}```
""",
    LLMProvider.cohere: """
** Generate {questions_per_chunk} questions based on the following context.**
- Avoid simple keyword-based questions; instead, focus on deeper understanding, implications, or connections within the text.
- Aim for a mix of factual, analytical, and open-ended questions.

*** Strictly Adhere to the following output format ***
```
what is the mentinoed product ID?
Are there any preferred geographical locations for the product?
```

*** Don't enumerate, itemize or bullet the questions.***

context:
```{chunk_text}```
""",
    LLMProvider.onprem: """
** Generate {questions_per_chunk} questions based on the following context.**
- Avoid simple keyword-based questions; instead, focus on deeper understanding, implications, or connections within the text.
- Aim for a mix of factual, analytical, and open-ended questions.

*** Strictly Adhere to the following output format ***
```
what is the mentinoed product ID?
Are there any preferred geographical locations for the product?
```

*** Don't enumerate, itemize or bullet the questions.***

context:
```{chunk_text}```
""",
}


system_prompt = {
    LLMProvider.openai: """** You are an AI model designed to generate insightful questions from a given text.**""",
    LLMProvider.cohere: """** You are an AI model designed to generate insightful questions from a given text.**""",
    LLMProvider.onprem: """** You are an AI model designed to generate insightful questions from a given text.**""",
}


class OpenAIResponse(BaseModel):
    questions: List[str]
