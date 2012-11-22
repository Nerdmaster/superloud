require "rubygems"
require "rspec"
require "yaml"
require "ostruct"
require File.dirname(__FILE__) + '/../data/messages'

describe "Messages" do
  before(:each) do
    @messages = Louds::Data::Messages.instance
    @data = {
      "FIRST" => Louds::Data::Message.new("FIRST", "Somebody"),
      "SECOND" => Louds::Data::Message.new("SECOND", "Another Person")
    }
    YAML.stub(:load_file => @data)
    @messages.load

    # Alias private data for easier testing
    @rnd = @messages.instance_variable_get("@random_messages")
    @msg = @messages.instance_variable_get("@messages")
  end

  describe "#random" do
    it "should pop a message from the random array and return it" do
      @rnd.should_receive(:pop).once.and_return("FIRST")
      @messages.random.should eq(@data["FIRST"])
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
