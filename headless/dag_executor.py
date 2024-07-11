import logging

import networkx as nx
import yaml

logging.basicConfig(level=logging.INFO)


class DAGExecutor:
    def __init__(self, function_registry):
        self.dags = {}
        self.outputs = {}
        self.variables = {}
        self.function_registry = function_registry

    def load_dags_from_file(self, file_path):
        with open(file_path, "r") as file:
            config = yaml.safe_load(file)

        self.variables = config.get("variables", {})

        for dag_name, tasks in config.items():
            if dag_name == "variables":
                continue
            self.add_dag(dag_name)
            for task_name, task_info in tasks.items():
                func_name = task_info.get("function", task_name)
                func = self.function_registry.get(func_name)
                self.add_task(
                    dag_name,
                    task_name,
                    func,
                    task_info["dependencies"],
                    task_info.get("params", {}),
                )

    def add_dag(self, dag_name):
        if dag_name not in self.dags:
            self.dags[dag_name] = nx.DiGraph()

    def add_task(self, dag_name, task_name, func, dependencies=None, params=None):
        self.add_dag(dag_name)
        self.dags[dag_name].add_node(task_name, func=func, params=params)
        if dependencies:
            for dep in dependencies:
                self.dags[dag_name].add_edge(dep, task_name)

    def get_execution_order(self, dag_name):
        graph = self.dags[dag_name]
        try:
            order = list(nx.topological_sort(graph))
            return order
        except nx.NetworkXUnfeasible:
            raise Exception(f"The DAG '{dag_name}' has cycles, which is not allowed.")

    def get_task_func(self, dag_name, task_name):
        return self.dags[dag_name].nodes[task_name]["func"]

    def get_task_params(self, dag_name, task_name):
        return self.dags[dag_name].nodes[task_name]["params"]

    def execute_task(self, dag_name, task_name):
        logging.info(f"Executing task '{task_name}' in DAG '{dag_name}'")
        task_func = self.get_task_func(dag_name, task_name)
        task_params = self.get_task_params(dag_name, task_name)
        if task_func:
            inputs = {}
            for param, source in task_params.items():
                if source == "variable":
                    inputs[param] = self.variables[param]
                elif source in self.outputs:
                    inputs[param] = self.outputs[source]
            self.outputs[task_name] = task_func(inputs)
        logging.info(f"Finished executing task '{task_name}' in DAG '{dag_name}'")

    def execute_dag(self, dag_name):
        logging.info(f"Starting execution of DAG '{dag_name}'")
        order = self.get_execution_order(dag_name)
        for task_name in order:
            self.execute_task(dag_name, task_name)
        logging.info(f"Finished execution of DAG '{dag_name}'")

    def execute_all(self):
        for dag_name in self.dags:
            self.execute_dag(dag_name)

    def set_variable(self, name, value):
        logging.info(f"Setting variable '{name}' to '{value}'")
        self.variables[name] = value

    def update_variables(self, new_vars):
        logging.info(f"Updating variables with '{new_vars}'")
        self.variables.update(new_vars)
