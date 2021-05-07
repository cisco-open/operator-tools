---
title: Override parameters
weight: 10
---

This operator embeds certain Kubernetes types into its own custom resource definitions. That way you can configure resources and override settings for resources that the operator creates.

The types (for example, PodSpec, see [Overrides](override/) for the complete list) are the structural equivalent of their original Kubernetes counterparts, but the required fields are declared as optional using `omitempty`, and in certain cases some fields are omitted. The advantage of this method is that you have to add only the parameters you want to override to the custom resource of the operator, the Kubernetes defaults will be used automatically for everything else.

The custom resource descriptions list which parameters you can override. For detailed examples, see [Override examples](override-examples/).
