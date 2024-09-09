from argparse import ArgumentParser

from headless.utils import (
    fetch_job_definition,
    get_nomad_endpoint,
    restart_nomad_job,
    stop_nomad_job,
)

if __name__ == "__main__":
    parser = ArgumentParser()
    parser.add_argument(
        "--base-url",
        type=str,
        required=True,
        help="The api url of the local application like http://127.0.0.1:8000/api/.",
    )
    args = parser.parse_args()

    model_bazaar_job = "modelbazaar"

    nomad_endpoint = get_nomad_endpoint(args.base_url)

    definition = fetch_job_definition(
        job_id=model_bazaar_job, nomad_endpoint=nomad_endpoint
    )

    stop_nomad_job(model_bazaar_job, nomad_endpoint=nomad_endpoint)

    # Resuming the job. since the checkpointing was enabled, resuming is possible by the same function
    restart_nomad_job(nomad_endpoint, payload=definition)
