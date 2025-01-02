// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts/proxy/Clones.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "./YayoiCollection.sol";

/**
 * @title YayoiFactory
 * @notice Factory contract for creating and managing YayoiCollection NFT contracts
 */
contract YayoiFactory is Ownable {
    using Clones for address;
    using SafeERC20 for IERC20;

    /// @notice Implementation contract that will be cloned for each new collection
    YayoiCollection public collectionImpl;

    /// @notice Address where protocol fees are sent
    address public protocolFeeDestination;
    /// @notice Token used for payments (address(0) for ETH)
    IERC20 public paymentToken;
    /// @notice Price to create a new collection
    uint256 public creationPrice;
    /// @notice Base minimum bid price
    /// @dev It's important from an economic perspective to have a base value
    /// here, since there are associated costs with image generation and
    /// minting.
    uint256 public baseMinimumBidPrice;
    /// @notice Base auction duration
    /// @dev This should be set to a reasonable value so the agent does not
    /// have an excessive amount of mintings to do and the auctions remain
    /// competitive.
    uint64 public baseAuctionDuration;

    /// @notice Protocol fee percentage in basis points (10%)
    uint256 public constant PROTOCOL_FEE_BPS = 1000;

    /// @notice Tracks which addresses are authorized to sign minting requests
    mapping(address => bool) public isAuthorizedSigner;
    /// @notice Tracks which addresses are registered collection contracts
    mapping(address => bool) public registeredCollections;

    /// @notice Emitted when a new collection is created
    event CollectionCreated(address indexed collection, address indexed owner);
    /// @notice Emitted when a signer's authorization status changes
    event AuthorizedSignerUpdated(address indexed signer, bool isAuthorized);
    /// @notice Emitted when the payment token is updated
    event PaymentTokenUpdated(address indexed token);
    /// @notice Emitted when the creation price is updated
    event CreationPriceUpdated(uint256 price);
    /// @notice Emitted when the base minimum bid price is updated
    event BaseMinimumBidPriceUpdated(uint256 price);
    /// @notice Emitted when the base auction duration is updated
    event BaseAuctionDurationUpdated(uint64 duration);
    /// @notice Emitted when the protocol fee destination is updated
    event ProtocolFeeDestinationUpdated(address indexed destination);
    /// @notice Emitted when a system prompt URI hash is registered
    event SystemPromptUriHashRegistered(bytes32 indexed promptUriHash);

    /**
     * @notice Initializes the factory with payment settings and deploys implementation
     * @param _paymentToken Address of token used for payments
     * @param _creationPrice Price to create a new collection
     * @param _baseMinimumBidPrice Base minimum bid price
     * @param _baseAuctionDuration Base auction duration
     * @param _protocolFeeDestination Address to receive protocol fees
     */
    constructor(
        address _paymentToken,
        uint256 _creationPrice,
        uint256 _baseMinimumBidPrice,
        uint64 _baseAuctionDuration,
        address _protocolFeeDestination
    ) Ownable(msg.sender) {
        paymentToken = IERC20(_paymentToken);
        creationPrice = _creationPrice;
        baseMinimumBidPrice = _baseMinimumBidPrice;
        baseAuctionDuration = _baseAuctionDuration;
        protocolFeeDestination = _protocolFeeDestination;

        collectionImpl = new YayoiCollection();
    }

    /**
     * @notice Parameters for creating a new collection
     * @param name Collection name
     * @param symbol Collection symbol
     * @param systemPromptUri URI of the system prompt
     * @param paymentToken Address of token used for payments
     * @param minimumBidPrice Minimum bid price
     * @param auctionDuration Duration of the auction
     */
    struct CreateCollectionParams {
        string name;
        string symbol;
        string systemPromptUri;
        address paymentToken;
        uint256 minimumBidPrice;
        uint64 auctionDuration;
    }

    /**
     * @notice Creates a new collection with the given parameters
     * @param params Collection creation parameters
     * @return collection Address of the created collection
     */
    function createCollection(CreateCollectionParams memory params)
        external
        payable
        returns (address payable collection)
    {
        require(params.auctionDuration >= baseAuctionDuration, "Less than base auction duration");

        if (address(paymentToken) != address(0)) {
            paymentToken.safeTransferFrom(msg.sender, address(this), creationPrice);
        } else {
            require(msg.value >= creationPrice, "Insufficient payment");
        }

        bytes32 salt = keccak256(bytes(params.systemPromptUri));

        collection = payable(address(collectionImpl).cloneDeterministic(salt));

        YayoiCollection(collection).initialize(
            YayoiCollection.InitializeParams({
                name: params.name,
                symbol: params.symbol,
                factory: address(this),
                owner: msg.sender,
                systemPromptUri: params.systemPromptUri,
                paymentToken: params.paymentToken,
                minimumBidPrice: params.minimumBidPrice,
                auctionDuration: params.auctionDuration
            })
        );

        registeredCollections[address(collection)] = true;

        emit CollectionCreated(collection, msg.sender);
    }

    /**
     * @notice Updates whether an address is authorized to sign minting requests
     * @param signer Address to update
     * @param authorized New authorization status
     */
    function updateAuthorizedSigner(address signer, bool authorized) external onlyOwner {
        require(signer != address(0), "Invalid signer");

        isAuthorizedSigner[signer] = authorized;
        emit AuthorizedSignerUpdated(signer, authorized);
    }

    /**
     * @notice Updates the implementation contract used for new collections
     * @param _implementation Address of new implementation
     */
    function setImplementation(address payable _implementation) external onlyOwner {
        require(_implementation != address(0), "Invalid implementation");

        collectionImpl = YayoiCollection(_implementation);
    }

    /**
     * @notice Updates the token used for payments
     * @param token Address of new payment token
     */
    function setPaymentToken(address token) external onlyOwner {
        paymentToken = IERC20(token);
        emit PaymentTokenUpdated(token);
    }

    /**
     * @notice Updates the price to create a new collection
     * @param price New creation price
     */
    function setCreationPrice(uint256 price) external onlyOwner {
        creationPrice = price;
        emit CreationPriceUpdated(price);
    }

    /**
     * @notice Updates the base minimum bid price
     * @param price New base minimum bid price
     */
    function setBaseMinimumBidPrice(uint256 price) external onlyOwner {
        baseMinimumBidPrice = price;
        emit BaseMinimumBidPriceUpdated(price);
    }

    /**
     * @notice Updates the base auction duration
     * @param duration New base auction duration
     */
    function setBaseAuctionDuration(uint64 duration) external onlyOwner {
        baseAuctionDuration = duration;
        emit BaseAuctionDurationUpdated(duration);
    }

    /**
     * @notice Updates the address that receives protocol fees
     * @param destination New fee destination address
     */
    function setProtocolFeeDestination(address destination) external onlyOwner {
        require(destination != address(0), "Invalid destination");

        protocolFeeDestination = destination;
        emit ProtocolFeeDestinationUpdated(destination);
    }

    /**
     * @notice Checks if an address is a registered collection
     * @param collection Address to check
     * @return bool True if address is a registered collection
     */
    function isRegisteredCollection(address collection) public view returns (bool) {
        return registeredCollections[collection];
    }

    /**
     * @notice Gets a collection address from a system prompt URI
     * @param systemPromptUri System prompt URI
     * @return collection Address of the collection
     */
    function getCollectionFromSystemPromptUri(string memory systemPromptUri) public view returns (address) {
        bytes32 salt = keccak256(bytes(systemPromptUri));
        address collection = address(collectionImpl).predictDeterministicAddress(salt);

        require(registeredCollections[collection], "Collection not found");

        return collection;
    }

    /**
     * @notice Withdraws tokens or ETH from the contract
     * @param token Address of token to withdraw (address(0) for ETH)
     */
    function sweepTokens(address token) external onlyOwner {
        if (token == address(0)) {
            (bool success,) = msg.sender.call{value: address(this).balance}("");
            require(success, "ETH transfer failed");
        } else {
            IERC20(token).safeTransfer(msg.sender, IERC20(token).balanceOf(address(this)));
        }
    }

    /// @notice Allows the contract to receive ETH
    receive() external payable {}
}
