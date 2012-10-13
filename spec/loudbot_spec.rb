require "rubygems"
require "rspec"
require "yaml"
require File.dirname(__FILE__) + '/../loudbot'
require File.dirname(__FILE__) + '/../data/messages'

describe "loudbot.rb" do
  describe "#init_data" do
    before(:each) do
      # Hack fake yaml data
      File.stub(:exists? => true)
      YAML.stub(:load_file => {"FIRST" => 1, "SECOND" => 12})
    end

    it "should retrieve messages" do
      @messages = nil
      init_data

      @messages.should be_kind_of(Louds::Data::Messages)

      @messages.instance_variable_get("@messages").should eq({"FIRST" => 1, "SECOND" => 12})
      @messages.instance_variable_get("@random_messages").sort.should eq(["FIRST", "SECOND"])
    end
  end
end
