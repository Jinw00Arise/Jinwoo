-- tutoChatNPC portal script
-- Tutorial portal that asks if the player wants to skip to Lith Harbor

if portal.askYesNo("Would you like to skip the tutorials and head straight to Lith Harbor?") then
    log(player.getName() .. " is skipping tutorial to Lith Harbor")
    player.warp(104000000, "sp")  -- Lith Harbor
else
    portal.sayNext("Enjoy your trip through the tutorial!")
    portal.block()  -- Prevent passage since they declined
end
