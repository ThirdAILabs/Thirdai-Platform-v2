import gradio as gr

# INTERACT WITH API #################################################################################


def signin(email, password):
    """Should return auth token"""
    # TODO: Replace mock implementation. See "login" in operations.ipynb
    return "Sign in is not implemented"

def pull_models_and_workflows(model_name_prefix, token):
    """Should return list of models and workflows as JSON"""
    # TODO: Replace mock implementation. See "list_models" in operations.ipynb
    models = ["Pull models is not implemented"]
    # We'll get to workflows later. Keep the mock implementation for now.
    workflows = ["Pull workflows is not implemented"]
    return models, workflows

def deploy(username, model_name, token):
    """Should return deployment status"""
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    return f"Deploy is not implemented"

def create_model(username, model_name, base_model_identifier, documents, token):
    """Should return ID of newly created model"""
    # TODO: Replace mock implementation.  See "Use Cases" document for details.
    return "Create model is not implemented"

def get_references(retriever_id, query, token):
    """Should return references from retrieval model"""
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    return [{"content": "Get references is not implemented"}]

async def generate_answer(guardrail_id, query, references, token):
    """Generates an answer to answer the query using provided references"""
    # TODO: Replace mock implementation. See "Use Cases" document for details.
    import time
    words = f"Generate answer is not implemented"
    for word in words.split():
        time.sleep(0.1)
        yield word + " "

def upvote_reference(model_id, query, reference_id, token):
    """Upvotes a reference"""
    # TODO: Implement actual upvoting logic
    return "Upvote reference is not implemented"

def associate_phrases(model_id, source_phrase, target_phrase, token):
    """Associates two phrases"""
    # TODO: Implement actual association logic
    return "Associate phrases is not implemented"

def pull_documents(model_id, token):
    """Should return list of documents as JSON"""
    # TODO: Replace mock implementation
    return ["Pull documents is not implemented"]

def insert_documents(model_id, documents, token):
    """Should insert a document and return its ID"""
    # TODO: Replace mock implementation
    return "Insert documents is not implemented"

def delete_document(model_id, document_id, token):
    """Should delete a document and return status"""
    # TODO: Replace mock implementation
    return "Delete documents is not implemented"


# UI COMPONENTS #################################################################################


def signin_tab(state):
    with gr.Column():
        email = gr.Textbox(label="Email")
        password = gr.Textbox(label="Password")
        signin_button = gr.Button("Sign In")
        status = gr.Textbox(label="Status")
    
    def process_signin(email, password, state):
        token = signin(email, password)
        return [token, {"token": token, **state}]
    
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

    def process_deploy(username, model_name, state):
        response = deploy(username, model_name, state['token'])
        return response

    deploy_button.click(process_deploy, inputs=[username, model_name, state], outputs=[deployment_status])

def create_tab(state):
    with gr.Column():
        username = gr.Textbox(label="Username")
        model_name = gr.Textbox(label="Model Name")
        base_model_identifier = gr.Textbox(label="Base Model ID")
        documents = gr.File(label="Documents to Index", file_count="multiple")
        create_button = gr.Button("Create")
        new_model_id = gr.Textbox(label="New Model ID")

    def process_create_model(username, model_name, base_model_identifier, documents, state):
        new_model_id = create_model(username, model_name, base_model_identifier, documents, state['token'])
        return new_model_id

    create_button.click(process_create_model, inputs=[username, model_name, base_model_identifier, documents, state], outputs=[new_model_id])


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
        documents = gr.File(label="Documents to Insert", file_count="multiple")
        insert_button = gr.Button("Submit")
        insert_status = gr.Textbox(label="Insert Status")

        gr.Markdown("## Delete Document")
        document_id = gr.Textbox(label="Document ID")
        delete_button = gr.Button("Submit")
        delete_status = gr.Textbox(label="Delete Status")

    def process_pull_documents(model_id, state):
        documents = pull_documents(model_id, state.get('token', ''))
        return documents

    def process_insert_documents(model_id, documents, state):
        result = insert_documents(model_id, documents, state.get('token', ''))
        return f"Document inserted with ID: {result}"

    def process_delete_document(model_id, document_id, state):
        result = delete_document(model_id, document_id, state.get('token', ''))
        return result

    pull_button.click(process_pull_documents, inputs=[model_id, state], outputs=[documents_json])
    insert_button.click(process_insert_documents, inputs=[model_id, documents, state], outputs=[insert_status])
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
