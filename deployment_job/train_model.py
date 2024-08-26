from thirdai import bolt

import thirdai
print(thirdai.__version__)

model = bolt.UniversalDeepTransformer(
            data_types={
                "source": bolt.types.text(),
                "target": bolt.types.token_tags(
                    tags=[
                        "FULL_NAME",
                        "EMAIL",
                        "ADDRESS",
                        "DOB",
                        "NATIONAL_ID",
                        "PASSPORT",
                        "VISA_PERMIT",
                        "DRIVER_LICENSE",
                        "VEHICLE_REG",
                        "LOCATION",
                        "PLACE_OF_BIRTH",
                        "MOTHERS_MAIDEN_NAME"
                    ], default_tag='O'
                ),
            },
            target="target",
        )

model.train("/share/pratyush/combined1_72K.csv")

model.save("/home/pratyush/pretrained_model_policy1/model.udt")

