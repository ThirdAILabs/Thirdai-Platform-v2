import argparse
import os
import subprocess
import sys


def parse_args():
    parser = argparse.ArgumentParser()
    parser.add_argument("--host", type=str, default="http://localhost:80")
    parser.add_argument("--email", type=str, required=True)
    parser.add_argument("--password", type=str, required=True)
    args = parser.parse_args()
    return args


def main(args):
    cwd = os.path.dirname(os.path.dirname(__file__))
    script = "stress_tests.end_to_end_stress_test"
    config_names = ["small-pdf", "large-csv", "many-files"]

    successes, failures = [], []

    for config in config_names:
        print(f"Starting test for config: {config}:\n\n")
        result = subprocess.run(
            f"python3 -m {script} --config {config} --host {args.host} --email {args.email} --password {args.password}",
            shell=True,
            cwd=cwd,
        )

        if result.returncode == 0:
            successes.append(config)
        else:
            failures.append(config)
            print(f"Error in {config}: {result.stderr}")

        print(f"\n\nFinished test for config {config}\n\n\n")

    print("\nThe following tests have passed:")
    print("\n".join(f"\t- {config}" for config in successes))
    print("\nThe following tests have failed:")
    print("\n".join(f"\t- {config}" for config in failures))

    if failures:
        sys.exit(1)


if __name__ == "__main__":
    args = parse_args()

    main(args)
