require "rubygems"
require "rspec"
require "ostruct"
require File.dirname(__FILE__) + '/../commands'

describe "commands" do
  before(:each) do
    @irc = double("Net::YAIL")
    @event = OpenStruct.new(:channel => "#ngs")
    @size_data = {}
  end

  describe "#size" do
    before(:each) do
      @size_data = {:one => {:size => 40, :nick => "Nerdmaster"}}
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

    it "should return size in CM and inches" do
      @irc.should_receive(:msg).with("#ngs", /40 CM/)
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
end
