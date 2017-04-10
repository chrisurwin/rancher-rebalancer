# rancher-rebalancer
A simple re-balancing service for Rancher cattle environments

### Info
This template contains a single alpine based image that performs simple balancing of a service across hosts in a Rancher cattle environment.

In some instances if there are host failures or maintenance you may end up with a host that can be running all, or the majority of containers, associated with a service.

What this rebalancer does is attempt to redistribute them more evenly across hosts in an environment to prevent service downtime.

Services must be labelled for rebalancing to take place. Each service that you want to be rebalanced must have a label of ```rebalance``` set to ```true```

