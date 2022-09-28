featureTriggers
===============

Features that do not generate game state directly and instead activate subfeatures when they trigger. The implementation is game specific and
reuse of code is mostly from utilising feature products.


Package dependencies
====================

This package is dependant on the feature and featureProduct packages.


Adding new features
===================

A feature must be registerd to be available to the feature system, same as feature products.

Implement the Feature interface functions Trigger, Serialize and Deserialize. Triggers do not require internal data.