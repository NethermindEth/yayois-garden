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

    mapping(bytes32 => bool) public usedSystemPromptUriHashes;
    mapping(address => bool) public isAuthorizedSigner;
    mapping(address => bool) public registeredCollections;

    event CollectionCreated(address indexed collection, address indexed owner);
    event AuthorizedSignerUpdated(address indexed signer, bool isAuthorized);
    event PaymentTokenUpdated(address indexed token);
    event CreationPriceUpdated(uint256 price);
    event ProtocolFeeDestinationUpdated(address indexed destination);
    event SystemPromptUriHashRegistered(bytes32 indexed promptUriHash);

    constructor(address _paymentToken, uint256 _creationPrice, address _protocolFeeDestination) Ownable(msg.sender) {
        paymentToken = IERC20(_paymentToken);
        creationPrice = _creationPrice;
        protocolFeeDestination = _protocolFeeDestination;

        collectionImpl = new YayoiCollection();
    }

    struct CreateCollectionParams {
        string name;
        string symbol;
        string systemPromptUri;
        address paymentToken;
        uint256 promptSubmissionPrice;
    }

    function createCollection(CreateCollectionParams memory params)
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

        YayoiCollection(collection).initialize(
            YayoiCollection.InitializeParams({
                name: params.name,
                symbol: params.symbol,
                factory: address(this),
                owner: msg.sender,
                systemPromptUri: params.systemPromptUri,
                paymentToken: params.paymentToken,
                promptSubmissionPrice: params.promptSubmissionPrice
            })
        );

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

    function isRegisteredCollection(address collection) public view returns (bool) {
        return registeredCollections[collection];
    }

    function isSystemPromptUriHashUsed(bytes32 promptUriHash) public view returns (bool) {
        return usedSystemPromptUriHashes[promptUriHash];
    }

    function registerSystemPromptUriHash(bytes32 promptUriHash) external returns (bool) {
        require(!usedSystemPromptUriHashes[promptUriHash], "Prompt URI hash already used");

        usedSystemPromptUriHashes[promptUriHash] = true;
        emit SystemPromptUriHashRegistered(promptUriHash);

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
