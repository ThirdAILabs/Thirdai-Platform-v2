import argparse
import os
import subprocess
import sys


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", type=str, default="http://localhost:80")
    parser.add_argument("--email", type=str)
    parser.add_argument("--password", type=str)
    args = parser.parse_args()
    return args


def main(args):
    folder = os.path.dirname(__file__)
    script_path = os.path.join(folder, "end_to_end_stress_test.py")
    config_names = ["small-pdf", "large-csv", "many-files"]

    successes, failures = [], []

    for config in config_names:
        result = subprocess.run(
            f"python3 {script_path} --config {config} --host {args.host} --email {args.email} --password {args.password}",
            capture_output=True,
            text=True,
            shell=True,
        )

        if result.returncode == 0:
            successes.append(config)
        else:
            failures.append(config)
            print(f"Error in {config}: {result.stderr}")

    print("\nThe following tests have passed:")
    print("\n".join(f"\t- {config}" for config in successes))
    print("\nThe following tests have failed:")
    print("\n".join(f"\t- {config}" for config in failures))

    if failures:
        sys.exit(1)


if __name__ == "__main__":
    args = parse_args()

    main(args)
