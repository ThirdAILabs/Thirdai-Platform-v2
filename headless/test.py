import argparse
import os
import sys

from headless import add_basic_args
from headless.configs import Config
from headless.dag_executor import DAGExecutor
from headless.functions import functions_registry, initialize_flow
from headless.utils import get_configs


def main():
    """
    Main function to run the DAG-based test suite.

    Parses command-line arguments, loads configurations, initializes the DAG executor,
    and runs the specified DAGs or tasks.
    """
    parser = argparse.ArgumentParser(description="Run DAG-based test suite.")
    add_basic_args(parser)
    parser.add_argument(
        "--dag-file",
        type=str,
        help="Path to the DAG YAML file",
        default=os.path.join(os.path.dirname(__file__), "dag_config.yaml"),
    )
    parser.add_argument("--dag", type=str, help="Name of the DAG to run")
    parser.add_argument(
        "--task", type=str, help="Name of the individual task to run within the DAG"
    )
    parser.add_argument("--all", action="store_true", help="Run all DAGs")
    parser.add_argument("--run-name", type=str, required=True, help="Name of the run")
    parser.add_argument("--sharded", action="store_true", help="Run sharded training")

    args = parser.parse_args()
    dag_executor = DAGExecutor(function_registry=functions_registry)
    dag_executor.load_dags_from_file(args.dag_file)

    dag_executor.update_variables(
        {
            "sharded": args.sharded,
            "run_name": args.run_name,
        }
    )

    initialize_flow(args.base_url, args.email, args.password)

    if args.all:
        dag_executor.execute_all()
    elif args.dag and args.task:
        dag_executor.execute_task(args.dag, args.task)
    elif args.dag:
        dag_executor.execute_dag(args.dag)
    else:
        print("Please specify either --dag, --task, or --all")
        sys.exit(1)


if __name__ == "__main__":
    main()
