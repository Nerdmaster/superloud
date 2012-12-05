module Louds
module Data

class Message
  attr_reader :author, :views, :score, :text, :uid
  attr_accessor :container

  include Comparable

  # Loads a message's attributes from a hash
  def initialize(data)
    @text = data[:text] || data["text"]
    @author = data[:author] || data["author"]
    @views = data[:views] || data["views"] || 0
    @score = data[:score] || data["score"] || 1
    @uid = data[:uid] || data["uid"]
  end

  def view!
    @views += 1
    changed!
  end

  def upvote!
    @score += 1
    changed!
  end

  def downvote!
    @score -= 1
    changed!
  end

  def changed!
    @container.dirty!(:changed => self)
  end

  # Converts all important attributes to a hash of data, primarily to ease exporting
  def to_hash
    return { :author => @author, :views => @views, :score => @score, :text => @text }
  end

  # Comparisons are somewhat meaningless, but they allow easier operations like == and simple
  # sorting by text
  def <=>(message)
    for field in [:text, :score, :views, :author]
      val = self.send(field) <=> message.send(field)
      return val unless val.zero?
    end

    return 0
  end
end

end
end
