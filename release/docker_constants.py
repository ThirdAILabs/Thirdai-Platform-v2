class ImageNames:
    def __init__(self, **kwargs):
        self.THIRDAI_PLATFORM_IMAGE_NAME = "thirdai_platform"
        for key, value in kwargs.items():
            setattr(self, key, value)

    def to_list(self):
        return list(self.__dict__.values())

    def peripherals_as_dict(self):
        as_dict = {attr: getattr(self, attr) for attr in self.__dict__}
        del as_dict["THIRDAI_PLATFORM_IMAGE_NAME"]
        return as_dict


# IMPORTANT: Image names are assumed to be the same as the directories that
# contain the docker files. E.g. TRAIN_IMAGE_NAME=train_job so the
# dockerfile that defines this image is assumed to be found in
# train_job/.
image_base_names = ImageNames(
    TRAIN_IMAGE_NAME="train_job",
    DEPLOY_IMAGE_NAME="deployment_job",
)
