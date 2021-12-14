# Advanced Hydrologic Prediction Service interface for Go

# About

This package provides an interface to the National Weather Service [Advanced Hydrologic Prediction Service](https://water.weather.gov/ahps2/)

## Todo List for 1.0
* Unexport any fields from the raw xml struct that we haven't decided on a stable api
* Cache layer, these data sets are only updated about once an hour. We can look at Site.Observed.Datum[0].Valid.Text and only pull it again when an hour has passed.
* Make the structs more user friendly. Things like Datum.Valid could have a better name and could be parsed into a time object.
