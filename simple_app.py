import gradio as gr


# INTERACT WITH API #################################################################################


def signin(username, password):
    # TODO: Replace mock implementation. See "login" in operations.ipynb
    if username == "admin" and password == "admin":
        return "token"
    else:
        return None

def pull_models_and_workflows(token):
    # TODO: Replace mock implementation. See "list_models" in operations.ipynb
    models = [
        {"id": "model1", "name": "Model 1"},
        {"id": "model2", "name": "Model 2"}
    ]
    # We'll get to workflows later. Keep the mock implementation for now.
    workflows = [
        {"id": "workflow1", "name": "Workflow 1"},
        {"id": "workflow2", "name": "Workflow 2"}
    ]
    return models, workflows

def create_model(model_name, base_model_id, documents, token):
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    new_model_id = f"{model_name}_{base_model_id}_{len(documents)}"
    return new_model_id

def get_references(retriever_id, query, token):
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    import random
    num_refs = random.randint(3, 10)
    return [{"content": f"Reference {i} for query: {query}. Retriever ID: {retriever_id}. Token: {token}"} for i in range(num_refs)]

def generate_answer(guardrail_id, references, token):
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    import time
    words = f"Token: {token}. Generated answer using guardrail {guardrail_id} based on {len(references)} references: "
    words += "This is a mock streaming response. " * 5
    for word in words.split():
        time.sleep(0.1)
        yield word + " "


# UI COMPONENTS #################################################################################


def signin_tab(state):
    with gr.Column():
        username = gr.Textbox(label="Username")
        password = gr.Textbox(label="Password")
        signin_button = gr.Button("Sign In")
        status = gr.Textbox(label="Status")
    
    def process_signin(username, password, state):
        token = signin(username, password)
        if token:
            status = "Sign-in successful"
        else:
            status = "Invalid credentials"
        return [status, {"token": token, **state}]
    
    signin_button.click(process_signin, inputs=[username, password, state], outputs=[status, state])


def models_tab(state):
    with gr.Column():
        pull_button = gr.Button("Pull models and workflows")
        models_json = gr.JSON(label="Models")
        workflows_json = gr.JSON(label="Workflows")

    def process_pull_models_and_workflows(state):
        return pull_models_and_workflows(state['token'])

    pull_button.click(process_pull_models_and_workflows, inputs=[state], outputs=[models_json, workflows_json])

def create_tab(state):
    with gr.Column():
        model_name = gr.Textbox(label="Model Name")
        base_model_id = gr.Textbox(label="Base Model ID")
        documents = gr.File(label="Documents to Index", file_count="multiple")
        create_button = gr.Button("Create")
        new_model_id = gr.Textbox(label="New Model ID")

    def process_create_model(model_name, base_model_id, documents, state):
        new_model_id = create_model(model_name, base_model_id, documents, state['token'])
        return new_model_id

    create_button.click(process_create_model, inputs=[model_name, base_model_id, documents, state], outputs=[new_model_id])


def interact_tab(state):
    max_num_references = 10
    with gr.Column():
        with gr.Accordion("Set connection"):
            retriever_id = gr.Textbox(label="Retriever ID")
            guardrail_id = gr.Textbox(label="LLM Guardrail ID")
            
        query = gr.Textbox(label="Query")
        submit_button = gr.Button("Submit")
        generated_answer = gr.Textbox(label="Generated Answer")
        reference_boxes = [gr.JSON(label=f"Reference {i+1}", visible=False) for i in range(max_num_references)]

    def process_query(retriever_id, query, state):
        references = get_references(retriever_id, query, state.get('token', ''))
        return [gr.JSON(value=r, visible=True) for r in references] + [gr.JSON(visible=False) for _ in range(max_num_references - len(references))]
        
    def process_generate_answer(guardrail_id, state, *reference_boxes):
        generated_so_far = ""
        for word in generate_answer(guardrail_id, reference_boxes, state.get('token', '')):
            generated_so_far += word
            yield generated_so_far

    query.submit(process_query, inputs=[retriever_id, query, state], outputs=reference_boxes).then(
        process_generate_answer, inputs=[guardrail_id, state, *reference_boxes], outputs=[generated_answer]
    )
    submit_button.click(process_query, inputs=[retriever_id, query, state], outputs=reference_boxes).then(
        process_generate_answer, inputs=[guardrail_id, state, *reference_boxes], outputs=[generated_answer]
    )
    

def create_app():
    with gr.Blocks() as app:
        gr.State({})
        state = gr.State({})

        with gr.Tabs():
            with gr.TabItem("Sign In"):
                signin_tab(state)
            with gr.TabItem("Models"):
                models_tab(state)
            with gr.TabItem("Create"):
                create_tab(state)
            with gr.TabItem("Interact"):
                interact_tab(state)

    return app

if __name__ == "__main__":
    app = create_app()
    app.launch()
