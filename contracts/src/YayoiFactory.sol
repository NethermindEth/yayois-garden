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

    /// @notice Protocol fee percentage in basis points (10%)
    uint256 public constant PROTOCOL_FEE_BPS = 1000;

    /// @notice Tracks which system prompt URI hashes have been used
    mapping(bytes32 => bool) public usedSystemPromptUriHashes;
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
    /// @notice Emitted when the protocol fee destination is updated
    event ProtocolFeeDestinationUpdated(address indexed destination);
    /// @notice Emitted when a system prompt URI hash is registered
    event SystemPromptUriHashRegistered(bytes32 indexed promptUriHash);

    /**
     * @notice Initializes the factory with payment settings and deploys implementation
     * @param _paymentToken Address of token used for payments
     * @param _creationPrice Price to create a new collection
     * @param _protocolFeeDestination Address to receive protocol fees
     */
    constructor(address _paymentToken, uint256 _creationPrice, address _protocolFeeDestination) Ownable(msg.sender) {
        paymentToken = IERC20(_paymentToken);
        creationPrice = _creationPrice;
        protocolFeeDestination = _protocolFeeDestination;

        collectionImpl = new YayoiCollection();
    }

    /**
     * @notice Parameters for creating a new collection
     * @param name Collection name
     * @param symbol Collection symbol
     * @param systemPromptUri URI of the system prompt
     * @param paymentToken Address of token used for payments
     * @param promptSubmissionPrice Price to submit a prompt
     */
    struct CreateCollectionParams {
        string name;
        string symbol;
        string systemPromptUri;
        address paymentToken;
        uint256 promptSubmissionPrice;
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
     * @notice Checks if a system prompt URI hash has been used
     * @param promptUriHash Hash to check
     * @return bool True if hash has been used
     */
    function isSystemPromptUriHashUsed(bytes32 promptUriHash) public view returns (bool) {
        return usedSystemPromptUriHashes[promptUriHash];
    }

    /**
     * @notice Registers a system prompt URI hash as used
     * @param promptUriHash Hash to register
     * @return bool True if registration was successful
     */
    function registerSystemPromptUriHash(bytes32 promptUriHash) external returns (bool) {
        require(!usedSystemPromptUriHashes[promptUriHash], "Prompt URI hash already used");

        usedSystemPromptUriHashes[promptUriHash] = true;
        emit SystemPromptUriHashRegistered(promptUriHash);

        return true;
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
