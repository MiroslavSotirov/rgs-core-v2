featureProducts
===============

Features that are added to the game state. When possible, these features should be made generic and resuable across many titles.


Package dependencies
=====================

This package is dependant on the feature package.


Adding new features
===================

A feature must be registerd to be available to the feature system. This can be done by static initiation of an unused varable

```
var _ feature.Factory = feature.RegisterFeature(FEATURE_ID_EXPANDING_WILD, func() feature.Feature { return new(ExpandingWild) })
```


Feature products should contain json serilizable data and therefore implement the DataPtr function.

```
func (f *ExpandingWild) DataPtr() interface{} {
	return &f.Data
}
```

They also need to implement the Feature interface functions Trigger, Serialize and Deserialize.
