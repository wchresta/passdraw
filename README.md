# passdraw

Passdraw is an algorithm for giving out event passes fairly.

## Problem statement

Large events, like [dance events](https://swingtzerland.com), sell hundreds of
event passes. When registration for these events open, they regularly sell out
very fast, sometimes within seconds.
During those registration openings, technical problems on either the server
or the client side, can lead to bad user experience. In very bad cases,
events even have adjusted the amount of passes they sold to recover from
registration crashes.

## Requirements

1. A fixed amount of `n` passes are to be assigned to a number `m` of users, where `n < m`. `m-n` users will not get a pass.
1. **Stress free registration**: The time of registration does not change the chances of a user getting a pass.
1. **Couples registration**: A pair of users can define a constraint; either both get a pass, or none get a pass.
1. **Fairness**: Passes are assigned fairly, where fair means:
    1. Users who do not define any constraints have probability `1 / n` to get a pass.
    1. Probability for a user to get a pass only depends on actions taken by that user.
    1. Two users with the same constraints have the same probability to get a pass.
1. **Recycling of canceled passes**: After passes have been distributed to users: Should a user cancel their registration, that users pass can be recycled.

## Design
### Registration

* Users are allowed to register and change their registration until a registration close time.
* A user `U` can define a dependency `U -> D`, meaning `U` only gets a pass if `D` gets a pass.
* Once registration is closed, the following algorithm declares who gets to get a pass.

### Backward algorithm

The following algorithm chooses who does *not* get a pass (refusals).
After at least `m-n` refusals have been found, the algorithm finishes.

