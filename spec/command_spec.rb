require "rubygems"
require "rspec"
require "ostruct"
require File.dirname(__FILE__) + '/../commands'

describe "commands" do
  describe "biggestdong" do
    before(:each) do
      @irc = double("Net::YAIL")
      @event = OpenStruct.new(:channel => "#ngs")
      @size_data = {}
    end

    it "should not name anybody when @size_data is empty" do
      @irc.should_receive(:msg).with("#ngs", /NO DONGS TODAY/)
      biggestdong(@event, {})
    end

    it "should return the winner when there are no ties" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"}
      }
      @irc.should_receive(:msg).with("#ngs", /NERDMASTER'S/)
      biggestdong(@event, {})
    end

    it "should return 'X and Y' in the case of a two-way tie" do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"}
      }
      @irc.should_receive(:msg).with("#ngs", /TIE BETWEEN DIALBOT AND NERDMASTER/)
      biggestdong(@event, {})
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

    it "should return 'X, Y, and Z' in the case of a three-way tie"do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"},
        :four => {:size => 6, :nick => "Hal"}
      }
      @irc.should_receive(:msg).with("#ngs", /TIE BETWEEN DIALBOT, HAL, AND NERDMASTER/)
      biggestdong(@event, {})
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
      biggestdong(@event, {})
    end

    it "should return 'W, X, Y, and Z' in the case of a three-way tie"do
      @size_data = {
        :one => {:size => 5, :nick => "loser"},
        :two => {:size => 6, :nick => "Nerdmaster"},
        :three => {:size => 6, :nick => "DialBOT"},
        :four => {:size => 6, :nick => "Hal"},
        :five => {:size => 6, :nick => "pizza"}
      }
      @irc.should_receive(:msg).with("#ngs", /DIALBOT, HAL, NERDMASTER, AND PIZZA/)
      biggestdong(@event, {})
    end
  end
end
