require "rubygems"
require "rspec"
require "ostruct"

require 'spec_helper'
lib 'commands'

describe "commands" do
  before(:each) do
    @irc = double("Net::YAIL")
    @event = OpenStruct.new(:channel => "#ngs")
    @size_data = {}
  end

  describe "#help" do
    it "should return generic help with no command-specific request" do
      @irc.should_receive(:msg).with("#ngs", /^I HAVE COMMANDS AND THEY ARE/)
      help(@event, [])
    end

    it "should return an error message if too many parameters are sent in" do
      @irc.should_receive(:msg).with("#ngs", /ONE COMMAND AT A TIME/)
      help(@event, ["HELP", "RPS"])
    end

    it "should return an error message if command-specific help is requested and no command exists" do
      @irc.should_receive(:msg).with("#ngs", /^!RP IS NOT A COMMAND/)
      help(@event, ["RP"])
    end

    it "should spit out command-specific help when valid" do
      @irc.should_receive(:msg).with("#ngs", /^!SIZEME: /)
      help(@event, ["SIZEME"])
    end

    it "should be an ass when requesting help help" do
      @irc.should_receive(:msg).with("#ngs", /^OH WOW YOU ARE SO META/)
      help(@event, ["HELP"])
    end

    it "should really be pulling all the help and command info from a config file and plugins or something"
  end

  describe "#size" do
    before(:each) do
      @size_data = {:one => {:size => 80, :nick => "Nerdmaster"}}
      @event.nick = "JealousGuy"
    end

    it "should call help if there is no user specified" do
      should_receive(:help).with(@event, ["SIZE"])
      size(@event, [])
    end

    it "should address the user if they are requesting their own size" do
      @event.nick = "Nerdmaster"
      @irc.should_receive(:msg).with("#ngs", /^HEY NERDMASTER/)
      size(@event, ["NERDMASTER"])
    end

    it "should return size in CM" do
      @irc.should_receive(:msg).with("#ngs", /40.0 CM/)
      size(@event, ["NERDMASTER"])
    end

    it "should return size in inches" do
      @irc.should_receive(:msg).with("#ngs", /15.7 INCHES/)
      size(@event, ["NERDMASTER"])
    end

    it "should return no data if name isn't found" do
      @size_data = {}
      @irc.should_not_receive(:msg).with("#ngs", /\d+ CM/)
      @irc.should_receive(:msg).with("#ngs", /NO DONG.*NERDMASTER/)
      size(@event, ["NERDMASTER"])
    end
  end

  describe "#sizeme" do
    it "should call the size command with the event's nickname as a parameter" do
      @event.nick = "Nerdmaster"
      should_receive(:size).with(@event, [@event.nick])
      sizeme(@event, [])
    end
  end

  describe "#biggestdong" do
    it "should not name anybody when @size_data is empty" do
      @irc.should_receive(:msg).with("#ngs", /NO DONGS TODAY/)
      biggestdong(@event, [])
    end

    it "should return the winner when there are no ties" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"}
      }
      @irc.should_receive(:msg).with("#ngs", /NERDMASTER'S/)
      biggestdong(@event, [])
    end

    it "should return 'X and Y' in the case of a two-way tie" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"}
      }
      @irc.should_receive(:msg).with("#ngs", /TIE BETWEEN DIALBOT AND NERDMASTER/)
      biggestdong(@event, [])
    end

    it "should use 'BOTH' to describe dongs in a two-way tie" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"}
      }
      @irc.should_receive(:msg).with("#ngs", /THEY'RE BOTH/)
      biggestdong(@event, {})
    end

    it "should use 'ALL' to describe dongs in a three-way tie" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"},
        :four => {:size => 6, :nick => "Hal"}
      }
      @irc.should_receive(:msg).with("#ngs", /THEY'RE ALL/)
      biggestdong(@event, {})
    end

    it "should return 'X, Y, and Z' in the case of a three-way tie" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"},
        :four => {:size => 6, :nick => "Hal"}
      }
      @irc.should_receive(:msg).with("#ngs", /TIE BETWEEN DIALBOT, HAL, AND NERDMASTER/)
      biggestdong(@event, [])
    end

    it "shouldn't confuse non-winning ties" do
      @size_data = {
        :xyzzy => {:size => 5, :nick => "loser"},
        :fuzzy => {:size => 5, :nick => "loser 2"},
        :one => {:size => 5, :nick => "loser 3"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"},
        :four => {:size => 6, :nick => "Hal"},
        :five => {:size => 7, :nick => "somedude"}
      }
      @irc.should_receive(:msg).with("#ngs", /SOMEDUDE'S/)
      biggestdong(@event, [])
    end

    it "should return 'W, X, Y, and Z' in the case of a four-way tie" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"},
        :four => {:size => 6, :nick => "Hal"},
        :five => {:size => 6, :nick => "pizza"}
      }
      @irc.should_receive(:msg).with("#ngs", /DIALBOT, HAL, NERDMASTER, AND PIZZA/)
      biggestdong(@event, [])
    end
  end

  describe "#dongrank" do
    it "should properly display rankings" do
      @size_data = {
        :xyzzy => {:size => 5, :nick => "loser"},
        :fuzzy => {:size => 5, :nick => "loser 2"},
        :one => {:size => 5, :nick => "loser 3"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"},
        :four => {:size => 6, :nick => "Hal"},
        :five => {:size => 7, :nick => "somedude"}
      }
      @irc.should_receive(:msg).with("#ngs", "IN FIRST PLACE WE HAVE SOMEDUDE; IN SECOND PLACE WE HAVE DIALBOT, HAL, AND NERDMASTER")
      dongrank(@event, [])
    end
  end

  describe "#fair_dong_size" do
    # Okay, seriously I'm not one to test that data is *exactly* a certain value for a certain
    # input - that can get very tedious and not really help the suite.  But here, I find it is
    # helpful just to know that the range is what I actually think it is, and to ensure that if
    # I modify the logic, the range isn't changing without my being aware of it.
    it "should return expected values" do
      # Worst roll = 12
      Dice.stub(:rand) do |sides|
        0
      end
      fair_dong_size.should eq(12)

      # Best roll = 66
      Dice.stub(:rand) do |sides|
        sides - 1
      end
      fair_dong_size.should eq(66)

      # Most commonly rolled = 31 (this test won't work if we call rand() on an even number!)
      Dice.stub(:rand) do |sides|
        (sides - 1) / 2
      end
      fair_dong_size.should eq(31)
    end
  end

  describe "#user_hash" do
    it "should return a set number on any platform" do
      message = OpenStruct.new(:user => "nerdmaster", :host => "nerdbucket.com")
      user_hash(message).should eq(333037843349749495676311809822862358690)
    end
  end
end
