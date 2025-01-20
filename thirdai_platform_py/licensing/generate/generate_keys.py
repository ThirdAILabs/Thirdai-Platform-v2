# key_generation.py

import os

from cryptography.hazmat.primitives import serialization
from cryptography.hazmat.primitives.asymmetric import rsa

licensing_dir = os.path.dirname(os.path.dirname(os.path.realpath(__file__)))

# Generate a key pair
private_key = rsa.generate_private_key(public_exponent=65537, key_size=2048)
public_key = private_key.public_key()

# Save the private key
with open(os.path.join(licensing_dir, "generate", "private_key.pem"), "wb") as f:
    f.write(
        private_key.private_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PrivateFormat.TraditionalOpenSSL,
            encryption_algorithm=serialization.NoEncryption(),
        )
    )

# Save the public key
with open(os.path.join(licensing_dir, "verify", "public_key.pem"), "wb") as f:
    f.write(
        public_key.public_bytes(
            encoding=serialization.Encoding.PEM,
            format=serialization.PublicFormat.SubjectPublicKeyInfo,
        )
    )

print("Keys generated and saved.")
