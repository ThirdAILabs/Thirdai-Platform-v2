from typing import List

from platform_common.pydantic_models.llm_config import LLMProvider
from pydantic import BaseModel

user_prompt = {
    LLMProvider.openai: """
** Generate {questions_per_chunk} high-quality questions based on the following context. **
- Ensure each question contains enough context to be answerable without requiring external knowledge.
- Avoid generic or vague questions that rely solely on keywords (e.g., `"What is the primary purpose of Figure 2?"`).
- Instead, focus on understanding, reasoning, and making meaningful connections within the text (e.g., `"What characteristics of in-text shape and layout make OCR challenging?"`).
- Include a mix of factual, analytical, and explanatory questions that probe deeper into the content.
- Refrain from open-ended, broad questions unless the context supports a well-defined answer.

context:
```{chunk_text}```
""",
    LLMProvider.cohere: """
** Generate {questions_per_chunk} questions based on the following context.**
- Ensure each question contains enough context to be answerable without requiring external knowledge.
- Avoid generic or vague questions that rely solely on keywords (e.g., `"What is the primary purpose of Figure 2?"`).
- Instead, focus on understanding, reasoning, and making meaningful connections within the text (e.g., `"What characteristics of in-text shape and layout make OCR challenging?"`).
- Include a mix of factual, analytical, and explanatory questions that probe deeper into the content.
- Refrain from open-ended, broad questions unless the context supports a well-defined answer.

*** Strictly Adhere to the following output format ***
```
what is the mentinoed product ID?
Are there any preferred geographical locations for the product?
```

*** 
- Don't enumerate, itemize or bullet the questions.
- Don't add any other prefix or suffix.
- Just output questions seperated by newline.
***

context:
```{chunk_text}```
""",
    LLMProvider.onprem: """
** Generate {questions_per_chunk} high-quality questions based on the following context. **
- Ensure each question contains enough context to be answerable without requiring external knowledge.
- Avoid generic or vague questions that rely solely on keywords (e.g., `"What is the primary purpose of Figure 2?"`).
- Instead, focus on understanding, reasoning, and making meaningful connections within the text (e.g., `"What characteristics of in-text shape and layout make OCR challenging?"`).
- Include a mix of factual, analytical, and explanatory questions that probe deeper into the content.
- Refrain from open-ended, broad questions unless the context supports a well-defined answer.

context:
```{chunk_text}```
""",
    LLMProvider.mock: """
** Generate {questions_per_chunk} high-quality questions based on the following context. **
- Ensure each question contains enough context to be answerable without requiring external knowledge.
- Avoid generic or vague questions that rely solely on keywords (e.g., `"What is the primary purpose of Figure 2?"`).
- Instead, focus on understanding, reasoning, and making meaningful connections within the text (e.g., `"What characteristics of in-text shape and layout make OCR challenging?"`).
- Include a mix of factual, analytical, and explanatory questions that probe deeper into the content.
- Refrain from open-ended, broad questions unless the context supports a well-defined answer.

context:
```{chunk_text}```
""",
}


system_prompt = {
    LLMProvider.openai: """** You are an AI model designed to generate insightful questions from a given text.**""",
    LLMProvider.cohere: """** You are an AI model designed to generate insightful questions from a given text.**""",
    LLMProvider.onprem: """** You are an AI model designed to generate insightful questions from a given text.**""",
    LLMProvider.mock: """** You are an AI model designed to generate insightful questions from a given text.**""",
}


class OpenAIResponse(BaseModel):
    questions: List[str]
