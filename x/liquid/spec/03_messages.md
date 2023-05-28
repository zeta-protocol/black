<!--
order: 3
-->

# Messages

`bblack` is minted using `MsgMintDerivative`.


```go
// MsgMintDerivative defines the Msg/MintDerivative request type.
type MsgMintDerivative struct {
	// sender is the owner of the delegation to be converted
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// validator is the validator of the delegation to be converted
	Validator string `protobuf:"bytes,2,opt,name=validator,proto3" json:"validator,omitempty"`
	// amount is the quantity of staked assets to be converted
	Amount types.Coin `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount"`
}
```

### Actions

* converts an existing delegation into bblack tokens
* delegation is transferred from the sender to a module account
* validator specific bblack are minted and sent to the sender

### Example:

```jsonc
{
  // user who owns the delegation
  "sender": "black10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t",
  // validator the user has delegated to
  "validator": "blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42",
  // amount of staked ufury to be converted into bblack
  "amount": {
    "amount": "1000000000",
    "denom": "ufury"
  }
}
```

`bblack` can be burned using `MsgBurnDerivative`.

```go
// MsgBurnDerivative defines the Msg/BurnDerivative request type.
type MsgBurnDerivative struct {
	// sender is the owner of the derivatives to be converted
	Sender string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	// validator is the validator of the derivatives to be converted
	Validator string `protobuf:"bytes,2,opt,name=validator,proto3" json:"validator,omitempty"`
	// amount is the quantity of derivatives to be converted
	Amount types.Coin `protobuf:"bytes,3,opt,name=amount,proto3" json:"amount"`
}
```

### Actions

* converts bblack tokens into a delegation
* bblack is burned
* a delegation equal to number of bblack is transferred to user


### Example

```jsonc
{
  // user who owns the bblack
  "sender": "black10wlnqzyss4accfqmyxwx5jy5x9nfkwh6qm7n4t",
  // the amount of bblack the user wants to convert back into normal staked black
  "amount": {
    "amount": "1234000000",
    "denom": "bblack-blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"
  },
  // the validator behind the bblack, this address must match the one embedded in the bblack denom above
  "validator": "blackvaloper1ypjp0m04pyp73hwgtc0dgkx0e9rrydeckewa42"
}
```
