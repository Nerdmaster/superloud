require "rubygems"
require "rspec"

require File.dirname(__FILE__) + '/../utils/utils'
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
