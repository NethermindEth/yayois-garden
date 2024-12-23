// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

import "forge-std/Test.sol";
import "../src/YayoiFactory.sol";
import "../src/YayoiCollection.sol";
import "./mocks/MockERC20.sol";

contract YayoiFactoryTest is Test {
    YayoiFactory public factory;
    MockERC20 public paymentToken;

    address public owner = address(1);
    address public signer = address(2);
    address public user = address(3);

    uint256 constant CREATION_PRICE = 1 ether;
    uint256 constant PROMPT_PRICE = 0.1 ether;

    function setUp() public {
        paymentToken = new MockERC20();
        factory = new YayoiFactory(address(paymentToken), CREATION_PRICE, address(this));

        vm.deal(user, 100 ether);
        paymentToken.transfer(user, 100 * 10 ** 18);

        vm.label(address(factory), "Factory");
        vm.label(address(paymentToken), "PaymentToken");
        vm.label(owner, "Owner");
        vm.label(signer, "Signer");
        vm.label(user, "User");
    }

    function testCreateCollection() public {
        YayoiFactory.CreateCollectionParams memory params = YayoiFactory.CreateCollectionParams({
            name: "Test Collection",
            symbol: "TEST",
            systemPromptUri: "ipfs://test",
            paymentToken: address(paymentToken),
            promptSubmissionPrice: PROMPT_PRICE
        });

        vm.startPrank(user);
        paymentToken.approve(address(factory), CREATION_PRICE);

        address predictedCollection = Clones.predictDeterministicAddress(
            address(factory.collectionImpl()), keccak256("ipfs://test"), address(factory)
        );

        vm.expectEmit(true, true, false, true);
        emit YayoiFactory.CollectionCreated(predictedCollection, user);

        address payable collection = factory.createCollection(params);
        vm.stopPrank();

        assertTrue(factory.isRegisteredCollection(collection));
        assertEq(YayoiCollection(collection).name(), "Test Collection");
        assertEq(YayoiCollection(collection).symbol(), "TEST");
        assertEq(YayoiCollection(collection).owner(), user);
    }

    function testFailCreateCollectionInsufficientPayment() public {
        YayoiFactory.CreateCollectionParams memory params = YayoiFactory.CreateCollectionParams({
            name: "Test Collection",
            symbol: "TEST",
            systemPromptUri: "ipfs://test",
            paymentToken: address(paymentToken),
            promptSubmissionPrice: PROMPT_PRICE
        });

        vm.prank(user);
        factory.createCollection(params);
    }

    function testUpdateAuthorizedSigner() public {
        factory.updateAuthorizedSigner(signer, true);
        assertTrue(factory.isAuthorizedSigner(signer));

        factory.updateAuthorizedSigner(signer, false);
        assertFalse(factory.isAuthorizedSigner(signer));
    }

    function testFailUpdateAuthorizedSignerUnauthorized() public {
        vm.prank(user);
        factory.updateAuthorizedSigner(signer, true);
    }

    function testSetImplementation() public {
        address newImpl = address(new YayoiCollection());
        factory.setImplementation(payable(newImpl));
        assertEq(address(factory.collectionImpl()), newImpl);
    }

    function testFailSetImplementationUnauthorized() public {
        vm.prank(user);
        factory.setImplementation(payable(address(1)));
    }

    function testSetPaymentToken() public {
        address newToken = address(new MockERC20());
        factory.setPaymentToken(newToken);
        assertEq(address(factory.paymentToken()), newToken);
    }

    function testSetCreationPrice() public {
        uint256 newPrice = 2 ether;
        factory.setCreationPrice(newPrice);
        assertEq(factory.creationPrice(), newPrice);
    }

    function testSetProtocolFeeDestination() public {
        address newDest = address(123);
        factory.setProtocolFeeDestination(newDest);
        assertEq(factory.protocolFeeDestination(), newDest);
    }

    function testRegisterSystemPrompt() public {
        YayoiFactory.CreateCollectionParams memory params = YayoiFactory.CreateCollectionParams({
            name: "Test Collection",
            symbol: "TEST",
            systemPromptUri: "test",
            paymentToken: address(paymentToken),
            promptSubmissionPrice: PROMPT_PRICE
        });

        vm.startPrank(user);
        paymentToken.approve(address(factory), CREATION_PRICE);
        factory.createCollection(params);
        vm.stopPrank();

        address collection = factory.getCollectionFromSystemPromptUri("test");
        assertTrue(factory.isRegisteredCollection(collection));
        assertEq(
            collection,
            Clones.predictDeterministicAddress(address(factory.collectionImpl()), keccak256("test"), address(factory))
        );
    }

    function testSweepTokens() public {
        // Send some tokens to contract
        paymentToken.transfer(address(factory), 1 ether);

        uint256 balanceBefore = paymentToken.balanceOf(address(this));
        factory.sweepTokens(address(paymentToken));
        uint256 balanceAfter = paymentToken.balanceOf(address(this));

        assertEq(balanceAfter - balanceBefore, 1 ether);
        assertEq(paymentToken.balanceOf(address(factory)), 0);
    }

    function testFailSweepTokensUnauthorized() public {
        vm.prank(user);
        factory.sweepTokens(address(paymentToken));
    }
}
