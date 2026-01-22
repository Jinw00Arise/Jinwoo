-- NPC Script: Maple Administrator (9010000)
-- A simple NPC that offers various services

npc.say("Hello, " .. player.getName() .. "! I'm the Maple Administrator.")
npc.say("Your current level is " .. player.getLevel() .. ".")
npc.say("You have " .. player.getMesos() .. " mesos.")

-- Note: Full NPC conversation support requires async/yield implementation
-- This is a basic example showing the available functions

npc.dispose()
