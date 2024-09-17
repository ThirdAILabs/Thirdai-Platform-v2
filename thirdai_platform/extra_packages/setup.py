from setuptools import setup, find_packages

setup(
    name="extra_packages",
    version="0.1",
    packages=find_packages(include=["thirdai_storage"]),
    install_requires=[
        "sqlalchemy",  # Add your dependencies here
        "pydantic",
    ],
    package_dir={"": "."},  # Adjust this to point to the extra_packages directory
)
