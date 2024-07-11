import argparse


def add_basic_args(parser: argparse.ArgumentParser):
    parser.add_argument(
        "--base-url",
        type=str,
        required=True,
        help="The api url of the local application like http://127.0.0.1:8000/api/.",
    )
    parser.add_argument("--email", type=str, required=True)
    parser.add_argument("--password", type=str, required=True)
