import base64
import json
import logging
import os
from datetime import datetime, timezone

import requests
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import padding
from cryptography.hazmat.primitives.serialization import load_pem_public_key
from thirdai import licensing

TASK_RUNNER_TOKEN = os.getenv("TASK_RUNNER_TOKEN")


def verify_license(license_path):
    licensing_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))

    # Load the public key
    with open(os.path.join(licensing_dir, "verify", "public_key.pem"), "rb") as f:
        public_key = load_pem_public_key(f.read())

    # Read the combined license
    with open(license_path, "r") as f:
        ndb_enterprise_license = json.load(f)

    # Extract the license information and signature
    license_info = ndb_enterprise_license["license"]
    encoded_signature = ndb_enterprise_license["signature"]

    # Decode the signature from Base64
    signature = base64.b64decode(encoded_signature)

    # Serialize the license information for verification
    license_str = json.dumps(license_info, separators=(",", ":"))

    # Verify the signature
    public_key.verify(
        signature, license_str.encode(), padding.PKCS1v15(), hashes.SHA256()
    )
    expiry_date = datetime.fromisoformat(
        license_info["expiryDate"].replace("Z", "+00:00")
    )
    if datetime.now(timezone.utc) > expiry_date:
        raise ValueError("License is expired.")

    return license_info


def valid_job_allocation(license_info, nomad_server_url, new_job_cpu_mhz=0):
    cpu_mhz_limit = int(license_info["cpuMhzLimit"])
    used_cpu_mhz = 0

    headers = {"X-Nomad-Token": TASK_RUNNER_TOKEN}

    allocations_response = requests.get(
        f"{nomad_server_url}/v1/allocations",
        params={"resources": True},
        headers=headers,
    ).json()

    # Calculate the used CPU resources by allocations
    for allocation in allocations_response:
        if allocation["ClientStatus"] == "running":
            for task_name, task in allocation["AllocatedResources"]["Tasks"].items():
                used_cpu_mhz += task["Cpu"]["CpuShares"]

    return used_cpu_mhz + new_job_cpu_mhz < cpu_mhz_limit


def activate_thirdai_license(thirdai_license):
    if thirdai_license.startswith("file "):
        logging.info("activating file based license")
        license_data = thirdai_license[len("file ") :]
        with open("./thirdai.license", "wb") as f:
            f.write(base64.b64decode(license_data))
        licensing.set_path("./thirdai.license")
    else:
        logging.info("activating key based license")
        licensing.activate(thirdai_license)


def verify_and_activate(license_path):
    license_info = verify_license(license_path)
    activate_thirdai_license(license_info["boltLicenseKey"])
