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
    address public user2 = address(4);

    uint256 constant SIGNER_PRIVATE_KEY = 0xdef1;
    address public signer = vm.addr(SIGNER_PRIVATE_KEY);

    uint256 constant CREATION_PRICE = 1 ether;
    uint256 constant MIN_BID_PRICE = 0.1 ether;
    uint64 constant AUCTION_DURATION = 1 days;

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
            minimumBidPrice: MIN_BID_PRICE,
            auctionDuration: AUCTION_DURATION
        });

        collection = YayoiCollection(payable(factory.createCollection(params)));

        // Setup test accounts
        vm.deal(user, 100 ether);
        vm.deal(user2, 100 ether);
        paymentToken.transfer(user, 100 * 10 ** 18);
        paymentToken.transfer(user2, 100 * 10 ** 18);

        vm.label(address(collection), "Collection");
        vm.label(address(factory), "Factory");
        vm.label(address(paymentToken), "PaymentToken");
        vm.label(owner, "Owner");
        vm.label(signer, "Signer");
        vm.label(user, "User");
        vm.label(user2, "User2");
    }

    function testInitialization() public view {
        assertEq(collection.name(), "Test Collection");
        assertEq(collection.symbol(), "TEST");
        assertEq(collection.owner(), address(this));
        assertEq(address(collection.factory()), address(factory));
        assertEq(collection.systemPromptUri(), "ipfs://test");
        assertEq(address(collection.paymentToken()), address(paymentToken));
        assertEq(collection.minimumBidPrice(), MIN_BID_PRICE);
        assertEq(collection.auctionDuration(), AUCTION_DURATION);
    }

    function testSuggestPrompt() public {
        uint256 currentAuctionId = collection.getCurrentAuctionId();

        vm.startPrank(user);
        paymentToken.approve(address(collection), MIN_BID_PRICE);

        vm.expectEmit(true, true, false, true);
        emit YayoiCollection.PromptAuctionStarted(currentAuctionId, block.timestamp);
        collection.suggestPrompt(currentAuctionId, "Test prompt", MIN_BID_PRICE);

        YayoiCollection.Auction memory auction = collection.getAuction(currentAuctionId);
        assertEq(auction.highestBidder, user);
        assertEq(auction.highestBid, MIN_BID_PRICE);
        assertEq(auction.prompt, "Test prompt");
        vm.stopPrank();
    }

    function testOutbidPreviousBid() public {
        uint256 currentAuctionId = collection.getCurrentAuctionId();
        uint256 firstBid = MIN_BID_PRICE;
        uint256 secondBid = MIN_BID_PRICE * 2;

        // First bid
        vm.startPrank(user);
        paymentToken.approve(address(collection), firstBid);
        collection.suggestPrompt(currentAuctionId, "First prompt", firstBid);
        vm.stopPrank();

        // Second bid
        vm.startPrank(user2);
        paymentToken.approve(address(collection), secondBid);
        collection.suggestPrompt(currentAuctionId, "Second prompt", secondBid);
        vm.stopPrank();

        YayoiCollection.Auction memory auction = collection.getAuction(currentAuctionId);
        assertEq(auction.highestBidder, user2);
        assertEq(auction.highestBid, secondBid);
        assertEq(auction.prompt, "Second prompt");
        assertEq(collection.pendingWithdrawals(user), firstBid);
    }

    function testFinishAuction() public {
        uint256 currentAuctionId = collection.getCurrentAuctionId();

        vm.startPrank(user);
        paymentToken.approve(address(collection), MIN_BID_PRICE);
        collection.suggestPrompt(currentAuctionId, "Test prompt", MIN_BID_PRICE);
        vm.stopPrank();

        // Fast forward past auction end
        vm.warp(block.timestamp + AUCTION_DURATION + 1);

        // Generate signature for minting
        bytes32 MINT_TYPEHASH = keccak256("Mint(address to,string uri)");
        string memory tokenUri = "ipfs://token1";

        bytes32 structHash = keccak256(abi.encode(MINT_TYPEHASH, user, keccak256(bytes(tokenUri))));
        bytes32 digest = collection.domainSeparator();
        digest = keccak256(abi.encodePacked("\x19\x01", digest, structHash));

        (uint8 v, bytes32 r, bytes32 s) = vm.sign(SIGNER_PRIVATE_KEY, digest);
        bytes memory signature = abi.encodePacked(r, s, v);

        uint256 protocolFeeBefore = paymentToken.balanceOf(address(factory));

        vm.expectEmit(true, true, false, true);
        emit YayoiCollection.PromptAuctionFinished(currentAuctionId, user, "Test prompt");
        collection.finishPromptAuction(currentAuctionId, tokenUri, signature);

        uint256 protocolFee = (MIN_BID_PRICE * 1000) / 10000; // 10% fee
        assertEq(paymentToken.balanceOf(address(factory)) - protocolFeeBefore, protocolFee);

        YayoiCollection.Auction memory auction = collection.getAuction(currentAuctionId);
        assertTrue(auction.finished);
        assertEq(collection.ownerOf(0), user);
        assertEq(collection.tokenURI(0), tokenUri);
    }

    function testRevertIfInvalidSignatureFinishAuction() public {
        uint256 currentAuctionId = collection.getCurrentAuctionId();

        vm.startPrank(user);
        paymentToken.approve(address(collection), MIN_BID_PRICE);
        collection.suggestPrompt(currentAuctionId, "Test prompt", MIN_BID_PRICE);
        vm.stopPrank();

        vm.warp(block.timestamp + AUCTION_DURATION + 1);

        bytes memory invalidSignature = new bytes(65);
        vm.expectRevert();
        collection.finishPromptAuction(currentAuctionId, "ipfs://token1", invalidSignature);
    }

    function testSetMinimumBidPrice() public {
        uint256 newPrice = 0.2 ether;

        collection.setMinimumBidPrice(newPrice);
        assertEq(collection.minimumBidPrice(), newPrice);
    }

    function testRevertIfUnauthorizedSetMinimumBidPrice() public {
        vm.prank(user);
        vm.expectRevert();
        collection.setMinimumBidPrice(0.2 ether);
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
