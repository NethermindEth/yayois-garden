// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "@openzeppelin/contracts-upgradeable/token/ERC721/ERC721Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC721/extensions/ERC721URIStorageUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
import "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/interfaces/IERC5267.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

import "./YayoiFactory.sol";

/**
 * @title YayoiCollection
 * @notice An upgradeable ERC721 contract for managing AI-generated art
 * collections. Supports EIP-712 signatures for minting and includes payment
 * functionality.
 */
contract YayoiCollection is
    Initializable,
    ERC721URIStorageUpgradeable,
    OwnableUpgradeable,
    IERC5267,
    EIP712Upgradeable
{
    using SafeERC20 for IERC20;

    /// @dev EIP-712 typehash for minting signatures
    bytes32 private constant MINT_TYPEHASH = keccak256("Mint(address to,string uri)");

    /// @notice Counter for token IDs
    uint256 public nextTokenId;

    /// @notice Reference to the factory contract that created this collection
    YayoiFactory public factory;

    /// @notice URI containing the system prompt used for this collection
    string public systemPromptUri;
    /// @notice Token used for payments (address(0) for ETH)
    address public paymentToken;
    /// @notice Minimum price to submit a prompt
    uint256 public minimumBidPrice;

    /// @notice Creation timestamp, used as the starting point for auctions
    uint64 public creationTimestamp;
    /// @notice Duration of each auction in seconds
    uint64 public auctionDuration;

    /// @notice Struct to store auction information
    struct Auction {
        bool finished;
        address highestBidder;
        uint256 highestBid;
        string prompt;
    }

    /// @notice Mapping from auction ID to auction info
    mapping(uint256 => Auction) internal _auctions;

    /// @notice Mapping to track user bids that can be withdrawn
    mapping(address => uint256) public pendingWithdrawals;

    /// @dev Protocol fee in basis points (10%)
    uint256 private constant PROTOCOL_FEE_BPS = 1000;

    /**
     * @notice Emitted when collection setup is completed
     * @param systemPromptUri The URI of the system prompt
     * @param paymentToken Address of the token used for payments
     * @param minimumBidPrice Minimum bid for auctions
     */
    event SetupCompleted(string systemPromptUri, address paymentToken, uint256 minimumBidPrice);

    /**
     * @notice Emitted when a prompt is submitted and token is minted
     * @param submitter Address that submitted the prompt
     * @param tokenId ID of the minted token
     * @param uri URI containing the prompt and generated art
     */
    event PromptSubmitted(address indexed submitter, uint256 indexed tokenId, string uri);

    /**
     * @notice Emitted when prompt submission price is updated
     * @param price New price for prompt submissions
     */
    event MinimumBidPriceUpdated(uint256 price);

    /**
     * @notice Emitted when a new auction starts
     * @param auctionId ID of the auction
     * @param startTime Start time of the auction
     */
    event PromptAuctionStarted(uint256 indexed auctionId, uint256 startTime);

    /**
     * @notice Emitted when a new bid is placed
     * @param auctionId ID of the auction
     * @param bidder Address of the bidder
     * @param amount Bid amount
     */
    event PromptAuctionBid(uint256 indexed auctionId, address indexed bidder, uint256 amount);

    /**
     * @notice Emitted when an auction is finished
     * @param auctionId ID of the auction
     * @param winner Address of the auction winner
     * @param prompt The winning prompt
     */
    event PromptAuctionFinished(uint256 indexed auctionId, address indexed winner, string prompt);

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Parameters for initializing the collection
     * @param name Collection name
     * @param symbol Collection symbol
     * @param factory Address of the factory contract
     * @param owner Address that will own the collection
     * @param systemPromptUri URI of the system prompt
     * @param paymentToken Address of token used for payments
     * @param minimumBidPrice Price to submit a prompt
     * @param auctionDuration Duration of each auction in seconds
     */
    struct InitializeParams {
        string name;
        string symbol;
        address factory;
        address owner;
        string systemPromptUri;
        address paymentToken;
        uint256 minimumBidPrice;
        uint64 auctionDuration;
    }

    /**
     * @notice Initializes the collection with the given parameters
     * @param params Initialization parameters
     */
    function initialize(InitializeParams memory params) public initializer {
        require(params.factory != address(0), "Invalid factory");
        require(params.auctionDuration > 0, "Invalid auction duration");

        __ERC721_init(params.name, params.symbol);
        __ERC721URIStorage_init();
        __Ownable_init(params.owner);
        __EIP712_init(params.name, "1");

        factory = YayoiFactory(payable(params.factory));
        systemPromptUri = params.systemPromptUri;
        paymentToken = params.paymentToken;
        minimumBidPrice = params.minimumBidPrice;
        auctionDuration = params.auctionDuration;
        creationTimestamp = uint64(block.timestamp);

        emit SetupCompleted(params.systemPromptUri, params.paymentToken, params.minimumBidPrice);
    }

    /**
     * @notice Gets the current auction ID based on current timestamp
     * @return auctionId The current auction ID
     */
    function getCurrentAuctionId() public view returns (uint256) {
        return (block.timestamp - creationTimestamp) / auctionDuration;
    }

    /**
     * @notice Gets the auction ID for a specific timestamp
     * @param timestamp The timestamp to get the auction ID for
     * @return auctionId The auction ID for that timestamp
     */
    function getAuctionIdByTimestamp(uint256 timestamp) public view returns (uint256) {
        return (timestamp - creationTimestamp) / auctionDuration;
    }

    /**
     * @notice Gets the start time of an auction
     * @param auctionId The auction ID
     * @return startTime The start time of the auction
     */
    function getAuctionStartTime(uint256 auctionId) public view returns (uint256) {
        return creationTimestamp + auctionId * auctionDuration;
    }

    /**
     * @notice Gets the end time of an auction
     * @param auctionId The auction ID
     * @return endTime The end time of the auction
     */
    function getAuctionEndTime(uint256 auctionId) public view returns (uint256) {
        return creationTimestamp + (auctionId + 1) * auctionDuration;
    }

    /**
     * @notice Suggests a prompt and starts an auction
     * @param auctionId The auction ID to submit for
     * @param prompt The prompt text to suggest
     * @param bid Bid amount
     */
    function suggestPrompt(uint256 auctionId, string memory prompt, uint256 bid) external payable {
        require(auctionId == getCurrentAuctionId(), "Invalid auction ID");
        require(bid >= minimumBidPrice, "Bid too low");

        Auction storage auction = _auctions[auctionId];
        if (auction.highestBidder == address(0)) {
            emit PromptAuctionStarted(auctionId, block.timestamp);
        }

        address previousBidder = auction.highestBidder;
        uint256 previousBid = auction.highestBid;

        require(bid > previousBid, "Not a winning bid");

        if (previousBid > 0) {
            pendingWithdrawals[previousBidder] += previousBid;
        }

        auction.highestBidder = msg.sender;
        auction.highestBid = bid;
        auction.prompt = prompt;

        emit PromptAuctionBid(auctionId, msg.sender, bid);

        uint256 userDepositedBalance = pendingWithdrawals[msg.sender];
        if (bid > userDepositedBalance) {
            pendingWithdrawals[msg.sender] = 0;
            _deposit(paymentToken, bid - userDepositedBalance);
        } else {
            pendingWithdrawals[msg.sender] = userDepositedBalance - bid;
        }
    }

    /**
     * @notice Finishes an auction after its duration has passed
     * @param auctionId The ID of the auction to finish
     */
    function finishPromptAuction(uint256 auctionId, string memory uri, bytes memory signature) external {
        require(block.timestamp >= getAuctionEndTime(auctionId), "Auction still active");
        Auction storage auction = _auctions[auctionId];
        require(!auction.finished, "Auction already finished");
        require(auction.highestBidder != address(0), "No bids placed");

        auction.finished = true;

        uint256 tokenId = nextTokenId;
        nextTokenId = tokenId + 1;

        // Verify signature
        bytes32 structHash = keccak256(abi.encode(MINT_TYPEHASH, auction.highestBidder, keccak256(bytes(uri))));
        bytes32 hash = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(hash, signature);
        require(factory.isAuthorizedSigner(signer), "Invalid signature");

        // Mint NFT
        _safeMint(auction.highestBidder, tokenId);
        _setTokenURI(tokenId, uri);

        // Calculate and transfer protocol fee
        uint256 protocolFee = (auction.highestBid * PROTOCOL_FEE_BPS) / 10000;
        _withdraw(paymentToken, address(factory), protocolFee);

        // Keep collection fee in contract
        emit PromptAuctionFinished(auctionId, auction.highestBidder, auction.prompt);
    }

    /**
     * @notice Withdraws pending refunds from outbid auctions
     */
    function withdrawPendingBids() external {
        uint256 amount = pendingWithdrawals[msg.sender];
        require(amount > 0, "No pending withdrawals");

        pendingWithdrawals[msg.sender] = 0;
        _withdraw(paymentToken, msg.sender, amount);
    }

    /**
     * @notice Sets a new price for prompt submissions
     * @param _price New price in payment token units
     */
    function setMinimumBidPrice(uint256 _price) external onlyOwner {
        minimumBidPrice = _price;
        emit MinimumBidPriceUpdated(_price);
    }

    /**
     * @notice Withdraws tokens or ETH from the contract
     * @param token Address of token to withdraw (address(0) for ETH)
     */
    function sweepTokens(address token) external onlyOwner {
        _withdrawAll(token, msg.sender);
    }

    /**
     * @notice Returns the auction information for a given auction ID
     * @param auctionId The ID of the auction
     * @return auction The auction information
     */
    function getAuction(uint256 auctionId) external view returns (Auction memory) {
        return _auctions[auctionId];
    }

    /**
     * @dev Transfer tokens or ETH from the contract
     * @param token Address of token to transfer (address(0) for ETH)
     * @param to Address to transfer to
     * @param amount Amount to transfer
     */
    function _withdraw(address token, address to, uint256 amount) internal {
        if (token == address(0)) {
            (bool success,) = to.call{value: amount}("");
            require(success, "ETH transfer failed");
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
    }

    /**
     * @dev Transfers all tokens or ETH from the contract to the owner
     */
    function _withdrawAll(address token, address to) internal onlyOwner {
        if (token == address(0)) {
            (bool success,) = to.call{value: address(this).balance}("");
            require(success, "ETH transfer failed");
        } else {
            IERC20(token).safeTransfer(to, IERC20(token).balanceOf(address(this)));
        }
    }

    /**
     * @dev Transfer tokens or ETH from sender to the contract
     * @param token Address of token to transfer (address(0) for ETH)
     * @param amount Amount to transfer
     */
    function _deposit(address token, uint256 amount) internal {
        if (token == address(0)) {
            require(msg.value >= amount, "Insufficient ETH");
        } else {
            IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        }
    }

    /**
     * @notice Returns the EIP-712 domain separator
     * @return bytes32 The domain separator
     */
    function domainSeparator() external view returns (bytes32) {
        return _domainSeparatorV4();
    }

    /// @notice Allows the contract to receive ETH
    receive() external payable {}
}
