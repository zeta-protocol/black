import { ethers } from "hardhat";

async function main() {
  const tokenName = "Black-wrapped ATOM";
  const tokenSymbol = "kATOM";
  const tokenDecimals = 6;

  const ERC20BlackWrappedCosmosCoin = await ethers.getContractFactory(
    "ERC20BlackWrappedCosmosCoin"
  );
  const token = await ERC20BlackWrappedCosmosCoin.deploy(
    tokenName,
    tokenSymbol,
    tokenDecimals
  );

  await token.deployed();

  console.log(
    `Token "${tokenName}" (${tokenSymbol}) with ${tokenDecimals} decimals is deployed to ${token.address}!`
  );
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
