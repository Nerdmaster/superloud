require "rubygems"
require "rspec"

require 'spec_helper'
lib 'data/messages'

describe "Messages" do
  before(:each) do
    @data = [
      {:text => "FIRST", :author => "Somebody", :views => 0, :score => 1},
      {:text => "SECOND", :author => "Another Person", :views => 0, :score => 1}
    ]
    @messages = Louds::Data::Messages.new("fakeyfake")
    @messages.stub(:retrieve_messages => @data)

    # Alias private data for easier testing
    @rnd = @messages.instance_variable_get("@random_messages")
    @msg = @messages.instance_variable_get("@messages")
    @voted = @messages.instance_variable_get("@voted")
  end

  describe "#load" do
    it "should convert the array from retrieve_messages into Message objects" do
      @msg.should eq({})
      @messages.load
      @msg.size.should eq(2)
      @msg["FIRST"].to_hash.should eq(@data[0])
      @msg["SECOND"].to_hash.should eq(@data[1])
    end
  end

  describe "#random" do
    before(:each) do
      # Auto-load messages for these tests
      @messages.load
    end

    it "should pop a message from the random array and return it" do
      @rnd.should_receive(:pop).once.and_return("FIRST")
      @messages.random.should eq(Louds::Data::Message.new(@data.first))
    end

    it "should reshuffle the messages array if random messages are empty" do
      @rnd.clear
      @msg.stub(:keys => [1, 2])
      @msg.keys.should_receive(:shuffle).once.and_return(["new stuff", "more new stuff"])

      @messages.random
    end

    it "should clear the list of voters" do
      @voted.should_receive(:clear).once
      @messages.random
    end
  end

  describe "#vote" do
    before(:each) do
      @last = double("Last message")
      @last.stub(:upvote!)
      @last.stub(:downvote!)

      @messages.instance_variable_set("@last", @last)
      @user_hash = 8473126312457
      @user_voted_hash = {@user_hash => true}
      @voted.clear
    end

    context "(when the user hasn't voted)" do
      it "should upvote on a vote of 1" do
        @last.should_receive(:upvote!).once
        @messages.vote(@user_hash, 1)
      end

      it "should downvote on a vote of -1" do
        @last.should_receive(:downvote!).once
        @messages.vote(@user_hash, -1)
      end

      it "shouldn't alter the score if the vote was anything but 1 or -1" do
        @last.should_not_receive(:downvote!)
        @last.should_not_receive(:upvote!)

        @messages.vote(@user_hash, -2)
        @messages.vote(@user_hash, 0)
        @messages.vote(@user_hash, 2)
      end

      it "should mark the user as having voted" do
        @voted.should eq({})
        @messages.vote(@user_hash, 1)
        @voted.should eq(@user_voted_hash)
      end

      it "should return true" do
        @messages.vote(@user_hash, 1).should eq(true)
      end
    end

    context "(when the user has already voted)" do
      before(:each) do
        @messages.vote(@user_hash, 1)
        @voted.should eq(@user_voted_hash)
      end

      it "shouldn't alter the message score" do
        @last.should_not_receive(:downvote!)
        @last.should_not_receive(:upvote!)
        @messages.vote(@user_hash, 1)
      end

      it "should still allow another person to vote" do
        @last.should_receive(:downvote!)
        @messages.vote(@user_hash + 1, -1)
      end

      it "should return false" do
        @messages.vote(@user_hash, 1).should eq(false)
      end
    end
  end
end

describe "Message" do
  describe "#add" do
    it "should store the new Message object in the hash" do
      pending
    end

    it "should mark the messages list dirty" do
      pending
    end
  end
end
