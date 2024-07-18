from typing import Any, Iterable, Literal, Union

from pydantic import BaseModel, Field, RootModel
from typing_extensions import Annotated

# https://docs.pydantic.dev/latest/concepts/unions/#discriminated-unions-with-str-discriminators
# This is the recommended way of polymorphism with validation in pydantic


class AnyOf(BaseModel):
    constraint_type: Literal["AnyOf"]
    values: Iterable[Any]


class EqualTo(BaseModel):
    constraint_type: Literal["EqualTo"]
    value: Any


class InRange(BaseModel):
    constraint_type: Literal["InRange"]
    minimum: Any
    maximum: Any
    inclusive_min: bool = True
    inclusive_max: bool = True


class GreaterThan(BaseModel):
    constraint_type: Literal["GreaterThan"]
    minimum: Any
    include_equal: bool = False


class LessThan(BaseModel):
    constraint_type: Literal["LessThan"]
    maximum: Any
    include_equal: bool = False


class Constraints(RootModel):
    root: dict[
        str,
        Annotated[
            Union[AnyOf, EqualTo, InRange, GreaterThan, LessThan],
            Field(discriminator="constraint_type"),
        ],
    ] = {}
