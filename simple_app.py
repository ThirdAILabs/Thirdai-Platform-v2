import gradio as gr
import operations as op


# INTERACT WITH API #################################################################################


def signin(email, password):
    """Should return auth token"""
    # TODO: Replace mock implementation. See "login" in operations.ipynb
    if email == "admin" and password == "admin":
        return "token"
    else:
        return None

def pull_models_and_workflows(model_name_prefix, token):
    """Should return list of models and workflows as JSON"""
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

def deploy(username, model_name, token):
    """Should return deployment status"""
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    return "Deployment not implemented"

def create_model(model_name, base_model_id, documents, token):
    """Should return ID of newly created model"""
    new_model_id = f"{model_name}_{base_model_id}_{len(documents)}"
    return new_model_id

def get_references(retriever_id, query, token):
    """Should return references from retrieval model"""
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    import random
    num_refs = random.randint(3, 10)
    return [{"content": f"Reference {i} for query: {query}. Retriever ID: {retriever_id}. Token: {token}"} for i in range(num_refs)]

async def generate_answer(guardrail_id, query, references, token):
    """Generates an answer to answer the query using provided references"""
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
        email = gr.Textbox(label="Email")
        password = gr.Textbox(label="Password")
        signin_button = gr.Button("Sign In")
        status = gr.Textbox(label="Status")
    
    def process_signin(email, password, state):
        token = signin(email, password)
        if token:
            status = "Sign-in successful"
        else:
            status = "Invalid credentials"
        return [status, {"token": token, **state}]
    
    signin_button.click(process_signin, inputs=[email, password, state], outputs=[status, state])


def models_tab(state):
    with gr.Column():
        model_name_prefix = gr.Textbox(label="Model Name Prefix")
        pull_button = gr.Button("Pull models and workflows")
        models_json = gr.JSON(label="Models")
        workflows_json = gr.JSON(label="Workflows")

    def process_pull_models_and_workflows(model_name_prefix, state):
        return pull_models_and_workflows(model_name_prefix, state['token'])

    pull_button.click(process_pull_models_and_workflows, inputs=[model_name_prefix, state], outputs=[models_json, workflows_json])

def deploy_tab(state):
    with gr.Column():
        username = gr.Textbox(label="Username")
        model_name = gr.Textbox(label="Model name")
        deploy_button = gr.Button("Deploy")
        deployment_status = gr.Textbox(label="Deployment Status")

    def process_deploy(username, modelname, state):
        response = deploy(username, model_name, state['token'])
        return response

    deploy_button.click(process_deploy, inputs=[username, model_name, state], outputs=[deployment_status])

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
        
    async def process_generate_answer(guardrail_id, state, query, *reference_boxes):
        generated_so_far = ""
        async for word in generate_answer(guardrail_id, query, reference_boxes, state.get('token', '')):
            generated_so_far += word
            yield generated_so_far

    query.submit(process_query, inputs=[retriever_id, query, state], outputs=reference_boxes).then(
        process_generate_answer, inputs=[guardrail_id, state, query, *reference_boxes], outputs=[generated_answer]
    )
    submit_button.click(process_query, inputs=[retriever_id, query, state], outputs=reference_boxes).then(
        process_generate_answer, inputs=[guardrail_id, state, query, *reference_boxes], outputs=[generated_answer]
    )
    

with gr.Blocks() as demo:
    gr.State({})
    state = gr.State({})

    with gr.Tabs():
        with gr.TabItem("Sign In"):
            signin_tab(state)
        with gr.TabItem("Models"):
            models_tab(state)
        with gr.TabItem("Deploy"):
            deploy_tab(state)
        with gr.TabItem("Create"):
            create_tab(state)
        with gr.TabItem("Interact"):
            interact_tab(state)


if __name__ == "__main__":
    demo.launch()
