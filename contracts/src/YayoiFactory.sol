// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/proxy/Clones.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./YayoiCollection.sol";

contract YayoiFactory is Ownable {
    using Clones for address;
    using SafeERC20 for IERC20;

    YayoiCollection public collectionImpl;

    address public protocolFeeDestination;
    IERC20 public paymentToken;
    uint256 public creationPrice;

    uint256 public constant PROTOCOL_FEE_BPS = 1000; // 10% fee in basis points

    mapping(bytes32 => bool) public usedPromptIds;
    mapping(address => bool) public isAuthorizedSigner;
    mapping(address => bool) public registeredCollections;

    event CollectionCreated(address indexed collection, address indexed owner);
    event AuthorizedSignerUpdated(address indexed signer, bool isAuthorized);
    event PaymentTokenUpdated(address indexed token);
    event CreationPriceUpdated(uint256 price);
    event ProtocolFeeDestinationUpdated(address indexed destination);
    event PromptIdRegistered(bytes32 indexed promptId);

    constructor(address _paymentToken, uint256 _creationPrice, address _protocolFeeDestination) Ownable(msg.sender) {
        paymentToken = IERC20(_paymentToken);
        creationPrice = _creationPrice;
        protocolFeeDestination = _protocolFeeDestination;

        collectionImpl = new YayoiCollection();
    }

    function createCollection(string calldata name, string calldata symbol)
        external
        payable
        returns (address payable collection)
    {
        if (address(paymentToken) != address(0)) {
            paymentToken.safeTransferFrom(msg.sender, address(this), creationPrice);
        } else {
            require(msg.value >= creationPrice, "Insufficient payment");
        }

        collection = payable(address(collectionImpl).clone());

        YayoiCollection(collection).initialize(name, symbol, address(this), msg.sender);
        registeredCollections[address(collection)] = true;

        emit CollectionCreated(collection, msg.sender);
    }

    function updateAuthorizedSigner(address signer, bool authorized) external onlyOwner {
        require(signer != address(0), "Invalid signer");

        isAuthorizedSigner[signer] = authorized;
        emit AuthorizedSignerUpdated(signer, authorized);
    }

    function setImplementation(address payable _implementation) external onlyOwner {
        require(_implementation != address(0), "Invalid implementation");

        collectionImpl = YayoiCollection(_implementation);
    }

    function setPaymentToken(address token) external onlyOwner {
        paymentToken = IERC20(token);
        emit PaymentTokenUpdated(token);
    }

    function setCreationPrice(uint256 price) external onlyOwner {
        creationPrice = price;
        emit CreationPriceUpdated(price);
    }

    function setProtocolFeeDestination(address destination) external onlyOwner {
        require(destination != address(0), "Invalid destination");

        protocolFeeDestination = destination;
        emit ProtocolFeeDestinationUpdated(destination);
    }

    function isPromptIdUsed(bytes32 promptId) public view returns (bool) {
        return usedPromptIds[promptId];
    }

    function isRegisteredCollection(address collection) public view returns (bool) {
        return registeredCollections[collection];
    }

    function registerPromptId(bytes32 promptId) external returns (bool) {
        require(!usedPromptIds[promptId], "Prompt ID already used");

        usedPromptIds[promptId] = true;
        emit PromptIdRegistered(promptId);

        return true;
    }

    function sweepTokens(address token) external onlyOwner {
        if (token == address(0)) {
            (bool success,) = msg.sender.call{value: address(this).balance}("");
            require(success, "ETH transfer failed");
        } else {
            IERC20(token).safeTransfer(msg.sender, IERC20(token).balanceOf(address(this)));
        }
    }

    receive() external payable {}
}
