import logging
from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import Any, Callable, Dict, List, Optional

import networkx as nx
import yaml

from headless.configs import Config
from headless.utils import get_configs

logging.basicConfig(level=logging.INFO)


class DAGExecutor:
    def __init__(self, function_registry: Dict[str, Callable]):
        """
        Initializes the DAGExecutor with a function registry.

        Parameters:
        function_registry (dict): A dictionary mapping function names to callable functions.
        """
        self.dags: Dict[str, nx.DiGraph] = {}
        self.outputs: Dict[str, Any] = {}
        self.variables: Dict[str, Any] = {}
        self.function_registry: Dict[str, Callable] = function_registry
        self.dag_configs: Dict[str, List[Config]] = {}

    def load_dags_from_file(self, file_path: str):
        """
        Loads DAGs from a YAML configuration file.

        Parameters:
        file_path (str): Path to the YAML file containing DAG configurations.
        """
        with open(file_path, "r") as file:
            config = yaml.safe_load(file)

        self.variables = config.get("variables", {})

        for dag_name, dag_info in config.items():
            if dag_name == "variables":
                continue
            self.add_dag(dag_name)
            config_names = dag_info.get("config")
            if isinstance(config_names, str):
                config_names = [config_names]
            self.dag_configs[dag_name] = [
                get_configs(Config, config_name)[0] for config_name in config_names
            ]
            for task_name, task_info in dag_info.items():
                if task_name == "config":
                    continue
                func_name = task_info.get("function", task_name)
                func = self.function_registry.get(func_name)
                self.add_task(
                    dag_name,
                    task_name,
                    func,
                    task_info.get("dependencies", []),
                    task_info.get("params", {}),
                )

    def add_dag(self, dag_name: str):
        """
        Adds a new DAG to the executor.

        Parameters:
        dag_name (str): Name of the DAG to add.
        """
        if dag_name not in self.dags:
            self.dags[dag_name] = nx.DiGraph()

    def add_task(
        self,
        dag_name: str,
        task_name: str,
        func: Callable,
        dependencies: Optional[List[str]] = None,
        params: Optional[Dict[str, str]] = None,
    ):
        """
        Adds a task to a specified DAG.

        Parameters:
        dag_name (str): Name of the DAG to add the task to.
        task_name (str): Name of the task.
        func (Callable): The function to execute for the task.
        dependencies (list, optional): List of task names that this task depends on.
        params (dict, optional): Dictionary of parameters for the task.
        """
        self.add_dag(dag_name)
        self.dags[dag_name].add_node(task_name, func=func, params=params)
        if dependencies:
            for dep in dependencies:
                self.dags[dag_name].add_edge(dep, task_name)

    def get_execution_order(self, dag_name: str) -> List[str]:
        """
        Returns the execution order of tasks in a DAG.

        Parameters:
        dag_name (str): Name of the DAG.

        Returns:
        list: A list of task names in the order they should be executed.

        Raises:
        Exception: If the DAG has cycles, which are not allowed.
        """
        graph = self.dags[dag_name]
        try:
            order = list(nx.topological_sort(graph))
            return order
        except nx.NetworkXUnfeasible:
            raise Exception(f"The DAG '{dag_name}' has cycles, which is not allowed.")

    def get_task_func(self, dag_name: str, task_name: str) -> Callable:
        """
        Retrieves the function associated with a task in a DAG.

        Parameters:
        dag_name (str): Name of the DAG.
        task_name (str): Name of the task.

        Returns:
        Callable: The function associated with the task.
        """
        return self.dags[dag_name].nodes[task_name]["func"]

    def get_task_params(self, dag_name: str, task_name: str) -> Dict[str, str]:
        """
        Retrieves the parameters associated with a task in a DAG.

        Parameters:
        dag_name (str): Name of the DAG.
        task_name (str): Name of the task.

        Returns:
        dict: The parameters associated with the task.
        """
        return self.dags[dag_name].nodes[task_name]["params"]

    def execute_task(self, dag_name: str, task_name: str, variables: Dict[str, Any]):
        """
        Executes a specified task in a DAG.

        Parameters:
        dag_name (str): Name of the DAG.
        task_name (str): Name of the task.
        variables (dict): Dictionary of variables for the task execution.
        """
        logging.info(f"Executing task '{task_name}' in DAG '{dag_name}'")
        task_func = self.get_task_func(dag_name, task_name)
        task_params = self.get_task_params(dag_name, task_name)
        if task_func:
            inputs = {}
            for param, source in task_params.items():
                if source == "variable":
                    inputs[param] = variables[param]
                elif source in self.outputs:
                    inputs[param] = self.outputs[source]
                else:
                    inputs[param] = source
            self.outputs[task_name] = task_func(inputs)
        logging.info(f"Finished executing task '{task_name}' in DAG '{dag_name}'")

    def execute_dag_with_config(self, dag_name: str, config: Config):
        """
        Executes a specified DAG with a given configuration.

        Parameters:
        dag_name (str): Name of the DAG.
        config (Config): Configuration to use for the DAG execution.
        """
        logging.info(
            f"Starting execution of DAG '{dag_name}' with config '{config.name}'"
        )
        variables = self.variables.copy()
        variables["config"] = config
        order = self.get_execution_order(dag_name)
        for task_name in order:
            self.execute_task(dag_name, task_name, variables)
        logging.info(
            f"Finished execution of DAG '{dag_name}' with config '{config.name}'"
        )

    def execute_dag(self, dag_name: str):
        """
        Executes all configurations of a specified DAG.

        Parameters:
        dag_name (str): Name of the DAG.
        """
        configs = self.dag_configs.get(dag_name)
        if not configs:
            raise ValueError(f"No configurations found for DAG '{dag_name}'")
        logging.info(
            f"Starting execution of DAG '{dag_name}' with {len(configs)} configurations"
        )
        for config in configs:
            self.execute_dag_with_config(dag_name, config)
        logging.info(f"Finished execution of DAG '{dag_name}' with all configurations")

    def execute_all(self):
        """
        Executes all DAGs managed by the executor, with all their configurations, in parallel.
        """
        logging.info(f"Starting execution of all DAGs")
        with ThreadPoolExecutor() as executor:
            futures = []
            for dag_name, configs in self.dag_configs.items():
                for config in configs:
                    futures.append(
                        executor.submit(self.execute_dag_with_config, dag_name, config)
                    )
            for future in as_completed(futures):
                try:
                    future.result()
                except Exception as exc:
                    logging.error(f"DAG generated an exception: {exc}")
        logging.info(f"Finished execution of all DAGs")

    def set_variable(self, name: str, value: Any):
        """
        Sets a variable to a specified value.

        Parameters:
        name (str): Name of the variable.
        value (Any): Value to set for the variable.
        """
        logging.info(f"Setting variable '{name}' to '{value}'")
        self.variables[name] = value

    def update_variables(self, new_vars: Dict[str, Any]):
        """
        Updates multiple variables with new values.

        Parameters:
        new_vars (dict): A dictionary of new variable values.
        """
        logging.info(f"Updating variables with '{new_vars}'")
        self.variables.update(new_vars)
