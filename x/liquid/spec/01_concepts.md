<!--
order: 1
-->

# Concepts

This module is responsible for the minting and burning of liquid staking receipt tokens, collectively referred to as `bblack`. Delegated black can be converted to delegator-specific `bblack`. Ie, 100 BLACK delegated to validator `blackvaloper123` can be converted to 100 `bblack-blackvaloper123`. Similarly, 100 `bblack-blackvaloper123` can be converted back to a delegation of 100 BLACK to  `blackvaloper123`. In this design, all validators can permissionlessly participate in liquid staking while users retain the delegator specific slashing risk and voting rights of their original validator. Note that because each `bblack` denom is validator specific, this module does not specify a fungibility mechanism for `bblack` denoms. 