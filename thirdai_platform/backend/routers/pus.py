import datetime
import os
import traceback
import uuid
from typing import List

import boto3
import pytz
from fastapi import APIRouter, Depends, status
from fastapi.responses import StreamingResponse
from database.session import get_session
from sqlalchemy.orm import Session
from database import schema

from auth.jwt import AuthenticatedUser, verify_access_token
from backend.utils import (
    State, 
    Task,
    TEXT_CLASSIFICATION_SUBTASKS,
    TOKEN_CLASSIFICATION_SUBTASKS,
    response,
    update_task_status,
    openai_client,
)

pus_router = APIRouter()
client = openai_client()

def find_occurences(input: str, item_list: List[str]):
    input = input.lower()
    return [item for item in item_list if item.lower() in input]

@pus_router.get("/infer-task")
def infer_task(
    problem_description: str,
    session: Session = Depends(get_session),
    authenticated_user: AuthenticatedUser = Depends(verify_access_token),
):

    # Starting a workflow
    workflow_id = uuid.uuid4()
    
    
    with get_session() as session:
        entry = schema.Workflow(id=workflow_id, user_id=authenticated_user.user.id)
        session.add(entry)
        session.commit()

    try:
        # Staring the task: infer task and sub-task
        infer_task_id = uuid.uuid4()
        
        with get_session() as session:
            # Adding the entry of the infer-task in the table
            entry = schema.InferTask(
                id=id, workflow_id=workflow_id, problem_description=problem_description
            )
            session.add(entry)

            # Updating the infer-task-status
            workflow_entry = (
                session.query(schema.Workflow).filter(schema.Workflow.id == workflow_id).first()
            )
            workflow_entry.infer_task_status = State.RUNNING
            session.commit()
            

        # Get the task from the openai
        prompt = f"""
            I will provide a customer's problem statement.
            Your job is to:
            1. Classify which of the following solution toolkits the customer could use.
            2. Provide why this solution toolkit can be useful to solve this problem.
            3. If none of the solution toolkits is feasible to solve the problem, simply return N/A.
            We support 2 different types of solution toolkits:
            1. Sentence classification (a ML model that translates a sentence to pre-defined categories)
            2. Token classification (a ML model that translates tokens within a sentence to pre-defined categories)
            Formulate your answer:
            1. Which solution toolkit to use.
            2. Why that toolkit would help solve user's problem.
            Notice that the toolkit can be a partial solution instead of a complete solution.
            Here is customer's problem statement: {problem_description}
            """

        # print("prompt", prompt)

        inferred_task = openai_client.openai_client(
            prompt=prompt,
            system_prompt="Act as a machine learning use case expert",
            # model_name="gpt-4o",
        )

        # print("inferred_task", inferred_task)
        inferred_task_value = None
        inferred_sub_task = None
        if "Sentence classification" in inferred_task:
            inferred_task_value = Task.TEXT_CLASSIFICATION
            all_supported_subtasks = "\n".join(TEXT_CLASSIFICATION_SUBTASKS)

            subtask_prompt = f"""
                We know that we are trying to solve a classification problem. Specifically, our solution is {inferred_task}.
                We need to further identify which exact type of classification it is.
                Here are a list of tasks we support: {all_supported_subtasks}\n\n
                Find at least 2 sub-tasks from this list which best suits the descriptions of the solution.
                Each sub-task should be output in separate lines.
                """

            # print("subtask_prompt", subtask_prompt)

            generated_inferred_subtask_text = client.llm_completion(
                prompt=subtask_prompt,
                # model_name="gpt-4o",
            )
            # print("inferred_sub_task", inferred_sub_task)
            inferred_sub_task = find_occurences(
                generated_inferred_subtask_text, item_list=TEXT_CLASSIFICATION_SUBTASKS
            )
        elif "Token classification" in inferred_task:
            inferred_task_value = Task.TOKEN_CLASSIFICATION
            inferred_sub_task = TOKEN_CLASSIFICATION_SUBTASKS
        else:
            raise ValueError(
                f"{inferred_task = }.\nGenerative Model didn't specify the inferred task"
            )

        # Completing the infer-tables row
        with get_session() as session:
            entry = session.query(schema.InferTask).filter(schema.InferTask.id == id).first()
            entry.task = (
                Task.TEXT_CLASSIFICATION.value
                if inferred_task_value == Task.TEXT_CLASSIFICATION
                else Task.TOKEN_CLASSIFICATION.value
            ),
            entry.sub_tasks=inferred_sub_task,

            with get_session() as session:
                entry = session.query(schema.Workflow).filter(schema.Workflow.id == workflow_id).first()
                entry.infer_task_status = status
                session.commit()
            session.commit()

        return response(
            status_code=status.HTTP_200_OK,
            message="Successfully inferred the task and sub-tasks",
            data={
                "task": inferred_task,
                "sub_tasks": inferred_sub_task,
                "workflow_id": str(workflow_id),
            },
        )
    except Exception as e:
        print(f"{workflow_id = }")
        return response(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            message="Got into an error. Unable to infer task. You can report back with this workflow id so that we can look into the issue",
            data={"workflow_id": str(workflow_id)},
        )