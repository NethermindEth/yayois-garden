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
    IERC20 public paymentToken;
    /// @notice Price to submit a prompt and mint a token
    uint256 public promptSubmissionPrice;

    /// @dev Protocol fee in basis points (10%)
    uint256 private constant PROTOCOL_FEE_BPS = 1000;

    /**
     * @notice Emitted when collection setup is completed
     * @param systemPromptUri The URI of the system prompt
     * @param paymentToken Address of the token used for payments
     * @param promptSubmissionPrice Price to submit a prompt
     */
    event SetupCompleted(string systemPromptUri, address paymentToken, uint256 promptSubmissionPrice);

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
    event PromptSubmissionPriceUpdated(uint256 price);

    /**
     * @notice Emitted when a prompt is suggested without minting
     * @param sender Address that suggested the prompt
     * @param prompt The suggested prompt text
     */
    event PromptSuggested(address indexed sender, string prompt);

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
     * @param promptSubmissionPrice Price to submit a prompt
     */
    struct InitializeParams {
        string name;
        string symbol;
        address factory;
        address owner;
        string systemPromptUri;
        address paymentToken;
        uint256 promptSubmissionPrice;
    }

    /**
     * @notice Initializes the collection with the given parameters
     * @param params Initialization parameters
     */
    function initialize(InitializeParams memory params) public initializer {
        require(params.factory != address(0), "Invalid factory");

        __ERC721_init(params.name, params.symbol);
        __ERC721URIStorage_init();
        __Ownable_init(params.owner);
        __EIP712_init(params.name, "1");

        factory = YayoiFactory(payable(params.factory));
        systemPromptUri = params.systemPromptUri;
        paymentToken = IERC20(params.paymentToken);
        promptSubmissionPrice = params.promptSubmissionPrice;

        emit SetupCompleted(params.systemPromptUri, params.paymentToken, params.promptSubmissionPrice);
    }

    /**
     * @notice Mints a new token with a verified signature
     * @param to Address to mint the token to
     * @param uri URI containing the prompt and generated art
     * @param signature EIP-712 signature from an authorized signer
     */
    function mintGeneratedToken(address to, string memory uri, bytes memory signature) external {
        require(address(factory) != address(0), "Not initialized");

        uint256 tokenId = nextTokenId;
        nextTokenId = tokenId + 1;

        // Verify signature
        bytes32 structHash = keccak256(abi.encode(MINT_TYPEHASH, to, keccak256(bytes(uri))));
        bytes32 hash = _hashTypedDataV4(structHash);
        address signer = ECDSA.recover(hash, signature);
        require(factory.isAuthorizedSigner(signer), "Invalid signature");

        // Mint NFT
        _safeMint(to, tokenId);
        _setTokenURI(tokenId, uri);

        emit PromptSubmitted(msg.sender, tokenId, uri);
    }

    /**
     * @notice Suggests a prompt without minting a token
     * @param prompt The prompt text to suggest
     */
    function suggestPrompt(string memory prompt) external payable {
        uint256 protocolFee = (promptSubmissionPrice * PROTOCOL_FEE_BPS) / 10000;
        uint256 collectionFee = promptSubmissionPrice - protocolFee;

        if (address(paymentToken) != address(0)) {
            paymentToken.safeTransferFrom(msg.sender, address(factory), protocolFee);
            paymentToken.safeTransferFrom(msg.sender, address(this), collectionFee);
        } else {
            require(msg.value >= promptSubmissionPrice, "Insufficient payment");

            (bool success1,) = payable(address(factory)).call{value: protocolFee}("");
            require(success1, "Protocol fee transfer failed");

            if (msg.value > promptSubmissionPrice) {
                (bool success2,) = payable(msg.sender).call{value: msg.value - promptSubmissionPrice}("");
                require(success2, "Excess ETH return failed");
            }
        }

        emit PromptSuggested(msg.sender, prompt);
    }

    /**
     * @notice Sets a new price for prompt submissions
     * @param _price New price in payment token units
     */
    function setPromptSubmissionPrice(uint256 _price) external onlyOwner {
        promptSubmissionPrice = _price;
        emit PromptSubmissionPriceUpdated(_price);
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
