import gradio as gr

# INTERACT WITH API #################################################################################


def signin(email, password):
    """Should return auth token"""
    # TODO: Replace mock implementation. See "login" in operations.ipynb
    if email == "admin" and password == "admin":
        return f"token_{email}"
    else:
        return None

def pull_models_and_workflows(model_name_prefix, token):
    """Should return list of models and workflows as JSON"""
    # TODO: Replace mock implementation. See "list_models" in operations.ipynb
    models = [
        {"id": f"model1_{model_name_prefix}", "name": f"Model 1 ({model_name_prefix})"},
        {"id": f"model2_{model_name_prefix}", "name": f"Model 2 ({model_name_prefix})"}
    ]
    # We'll get to workflows later. Keep the mock implementation for now.
    workflows = [
        {"id": f"workflow1_{model_name_prefix}", "name": f"Workflow 1 ({model_name_prefix})"},
        {"id": f"workflow2_{model_name_prefix}", "name": f"Workflow 2 ({model_name_prefix})"}
    ]
    return models, workflows

def deploy(username, model_name, token):
    """Should return deployment status"""
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    return f"Deployment not implemented for user {username}, model {model_name}, token {token}"

def create_model(model_name, base_model_id, documents, token):
    """Should return ID of newly created model"""
    new_model_id = f"{model_name}_{base_model_id}_{len(documents)}_{token}"
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
    words = f"Token: {token}. Generated answer using guardrail {guardrail_id} for query '{query}' based on {len(references)} references: "
    words += "This is a mock streaming response. " * 5
    for word in words.split():
        time.sleep(0.1)
        yield word + " "

def upvote_reference(model_id, query, reference_id, token):
    """Upvotes a reference"""
    # TODO: Implement actual upvoting logic
    return f"Upvoted reference {reference_id} for query: {query} in model {model_id}. Token: {token}"

def associate_phrases(model_id, source_phrase, target_phrase, token):
    """Associates two phrases"""
    # TODO: Implement actual association logic
    return f"Associated '{source_phrase}' with '{target_phrase}' in model {model_id}. Token: {token}"

def pull_documents(model_id, token):
    """Should return list of documents as JSON"""
    # TODO: Replace mock implementation
    documents = [
        {"id": f"doc1_{model_id}", "name": f"Document 1 for model {model_id}"},
        {"id": f"doc2_{model_id}", "name": f"Document 2 for model {model_id}"}
    ]
    return documents

def insert_document(model_id, document, token):
    """Should insert a document and return its ID"""
    # TODO: Replace mock implementation
    return f"doc_{model_id}_{document.name}_{token}"

def delete_document(model_id, document_id, token):
    """Should delete a document and return status"""
    # TODO: Replace mock implementation
    return f"Document {document_id} deleted from model {model_id}. Token: {token}"


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

def feedback_tab(state):
    with gr.Column():
        model_id = gr.Textbox(label="Model ID")
        
        gr.Markdown("## Upvote")
        query = gr.Textbox(label="Query")
        reference_id = gr.Number(label="Reference ID", precision=0)
        upvote_button = gr.Button("Submit")
        upvote_status = gr.Textbox(label="Upvote Status")

        gr.Markdown("## Associate")
        source_phrase = gr.Textbox(label="Source phrase")
        target_phrase = gr.Textbox(label="Target phrase")
        associate_button = gr.Button("Submit")
        associate_status = gr.Textbox(label="Association Status")

    def process_upvote(model_id, query, reference_id, state):
        result = upvote_reference(model_id, query, reference_id, state.get('token', ''))
        return result

    def process_associate(model_id, source_phrase, target_phrase, state):
        result = associate_phrases(model_id, source_phrase, target_phrase, state.get('token', ''))
        return result

    upvote_button.click(process_upvote, inputs=[model_id, query, reference_id, state], outputs=[upvote_status])
    associate_button.click(process_associate, inputs=[model_id, source_phrase, target_phrase, state], outputs=[associate_status])

def documents_tab(state):
    with gr.Column():
        model_id = gr.Textbox(label="Model ID")

        gr.Markdown("## List Documents")
        pull_button = gr.Button("Pull")
        documents_json = gr.JSON(label="Documents")

        gr.Markdown("## Insert Document")
        document = gr.File(label="Document to Insert")
        insert_button = gr.Button("Submit")
        insert_status = gr.Textbox(label="Insert Status")

        gr.Markdown("## Delete Document")
        document_id = gr.Textbox(label="Document ID")
        delete_button = gr.Button("Submit")
        delete_status = gr.Textbox(label="Delete Status")

    def process_pull_documents(model_id, state):
        documents = pull_documents(model_id, state.get('token', ''))
        return documents

    def process_insert_document(model_id, document, state):
        result = insert_document(model_id, document, state.get('token', ''))
        return f"Document inserted with ID: {result}"

    def process_delete_document(model_id, document_id, state):
        result = delete_document(model_id, document_id, state.get('token', ''))
        return result

    pull_button.click(process_pull_documents, inputs=[model_id, state], outputs=[documents_json])
    insert_button.click(process_insert_document, inputs=[model_id, document, state], outputs=[insert_status])
    delete_button.click(process_delete_document, inputs=[model_id, document_id, state], outputs=[delete_status])

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
        with gr.TabItem("Feedback"):
            feedback_tab(state)
        with gr.TabItem("Documents"):
            documents_tab(state)


if __name__ == "__main__":
    demo.launch()
