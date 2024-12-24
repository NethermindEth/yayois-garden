// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/YayoiCollection.sol";
import "../src/YayoiFactory.sol";
import "./mocks/MockERC20.sol";

contract YayoiCollectionTest is Test {
    YayoiCollection public collection;
    YayoiFactory public factory;
    MockERC20 public paymentToken;

    address public owner = address(1);
    address public user = address(3);

    uint256 constant SIGNER_PRIVATE_KEY = 0xdef1;
    address public signer = vm.addr(SIGNER_PRIVATE_KEY);

    uint256 constant CREATION_PRICE = 1 ether;
    uint256 constant PROMPT_PRICE = 0.1 ether;

    function setUp() public {
        // Deploy contracts
        paymentToken = new MockERC20();
        factory = new YayoiFactory(address(paymentToken), CREATION_PRICE, address(this));

        // Authorize signer
        factory.updateAuthorizedSigner(signer, true);

        // Create collection
        paymentToken.approve(address(factory), CREATION_PRICE);

        YayoiFactory.CreateCollectionParams memory params = YayoiFactory.CreateCollectionParams({
            name: "Test Collection",
            symbol: "TEST",
            systemPromptUri: "ipfs://test",
            paymentToken: address(paymentToken),
            promptSubmissionPrice: PROMPT_PRICE
        });

        collection = YayoiCollection(payable(factory.createCollection(params)));

        // Setup test accounts
        vm.deal(user, 100 ether);
        paymentToken.transfer(user, 100 * 10 ** 18);

        vm.label(address(collection), "Collection");
        vm.label(address(factory), "Factory");
        vm.label(address(paymentToken), "PaymentToken");
        vm.label(owner, "Owner");
        vm.label(signer, "Signer");
        vm.label(user, "User");
    }

    function testInitialization() public view {
        assertEq(collection.name(), "Test Collection");
        assertEq(collection.symbol(), "TEST");
        assertEq(collection.owner(), address(this));
        assertEq(address(collection.factory()), address(factory));
        assertEq(collection.systemPromptUri(), "ipfs://test");
        assertEq(address(collection.paymentToken()), address(paymentToken));
        assertEq(collection.promptSubmissionPrice(), PROMPT_PRICE);
    }

    function testMintWithValidSignature() public {
        // Generate signature
        bytes32 MINT_TYPEHASH = keccak256("Mint(address to,string uri)");
        string memory tokenUri = "ipfs://token1";

        bytes32 structHash = keccak256(abi.encode(MINT_TYPEHASH, user, keccak256(bytes(tokenUri))));

        bytes32 digest = collection.domainSeparator();
        digest = keccak256(abi.encodePacked("\x19\x01", digest, structHash));

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(SIGNER_PRIVATE_KEY, digest);
        bytes memory signature = abi.encodePacked(r, s, v);

        // Mint token
        vm.startPrank(user);
        paymentToken.approve(address(collection), PROMPT_PRICE);
        collection.mintGeneratedToken(user, tokenUri, signature);
        vm.stopPrank();

        assertEq(collection.ownerOf(0), user);
        assertEq(collection.tokenURI(0), tokenUri);
    }

    function testRevertIfInvalidSignature() public {
        bytes memory invalidSignature = new bytes(65);

        vm.startPrank(user);
        paymentToken.approve(address(collection), PROMPT_PRICE);
        vm.expectRevert();
        collection.mintGeneratedToken(user, "ipfs://token1", invalidSignature);
        vm.stopPrank();
    }

    function testSuggestPrompt() public {
        vm.startPrank(user);
        paymentToken.approve(address(collection), PROMPT_PRICE);

        uint256 protocolFeeBefore = paymentToken.balanceOf(address(factory));
        uint256 collectionFeeBefore = paymentToken.balanceOf(address(collection));

        vm.expectEmit(true, false, false, true);
        emit YayoiCollection.PromptSuggested(user, "Test prompt");
        collection.suggestPrompt("Test prompt");

        uint256 protocolFee = (PROMPT_PRICE * 1000) / 10000; // 10% fee
        uint256 collectionFee = PROMPT_PRICE - protocolFee;

        assertEq(paymentToken.balanceOf(address(factory)) - protocolFeeBefore, protocolFee);
        assertEq(paymentToken.balanceOf(address(collection)) - collectionFeeBefore, collectionFee);
        vm.stopPrank();
    }

    function testSetPromptSubmissionPrice() public {
        uint256 newPrice = 0.2 ether;

        collection.setPromptSubmissionPrice(newPrice);
        assertEq(collection.promptSubmissionPrice(), newPrice);
    }

    function testRevertIfUnauthorizedSetPromptSubmissionPrice() public {
        vm.prank(user);
        vm.expectRevert();
        collection.setPromptSubmissionPrice(0.2 ether);
    }

    function testSweepTokens() public {
        // Send some tokens to contract
        paymentToken.transfer(address(collection), 1 ether);

        uint256 balanceBefore = paymentToken.balanceOf(address(this));
        collection.sweepTokens(address(paymentToken));
        uint256 balanceAfter = paymentToken.balanceOf(address(this));

        assertEq(balanceAfter - balanceBefore, 1 ether);
        assertEq(paymentToken.balanceOf(address(collection)), 0);
    }

    function testRevertIfUnauthorizedSweepTokens() public {
        vm.prank(user);
        vm.expectRevert();
        collection.sweepTokens(address(paymentToken));
    }
}
