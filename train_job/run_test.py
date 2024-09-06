from run import get_model
from config import TrainConfig
from reporter import Reporter
from typing import Dict
import thirdai

class DummyReporter(Reporter):
    def report_complete(self, model_id: str, metadata: Dict[str, str]):
        pass

    def report_status(self, model_id: str, status: str, message: str = ""):
        pass


def main():
    with open("./test_config.json", "r") as file:
        config = TrainConfig.model_validate_json(file.read())

    thirdai.licensing.activate(config.license_key)

    model = get_model(config, DummyReporter())

    model.train()


if __name__ == "__main__":
    main()
