# frozen_string_literal: true
# SPDX-License-Identifier: BSD-3-Clause
require "json"
require_relative "_harness"
def build_doc
  objs = (0...60).map do |i|
    ok = i.even? ? "true" : "false"
    %Q({"id":#{i},"name":"item-#{i}","score":#{(i * 37) % 100},"ok":#{ok},"tags":[#{i},#{i * 2},#{i * 3}]})
  end
  "[" + objs.join(",") + "]"
end
doc  = build_doc
tree = JSON.parse(doc)
bench("parse-60obj",    1000) { JSON.parse(doc) }
bench("generate-60obj", 1000) { JSON.generate(tree) }
