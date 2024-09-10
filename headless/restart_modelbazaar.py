from headless.utils import fetch_job_definition, restart_nomad_job, stop_nomad_job

if __name__ == "__main__":
    model_bazaar_job = "modelbazaar"

    nomad_endpoint = "http://localhost:4646/"

    definition = fetch_job_definition(
        job_id=model_bazaar_job, nomad_endpoint=nomad_endpoint
    )

    stop_nomad_job(model_bazaar_job, nomad_endpoint=nomad_endpoint)

    # Resuming the job. since the checkpointing was enabled, resuming is possible by the same function
    restart_nomad_job(nomad_endpoint, payload=definition)
