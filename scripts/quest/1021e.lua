--[[
    Quest Script: q1021e
    Roger's Apple (1021 - end)
    NPC: Roger (2000)
]]

function start()
    -- Check if HP is fully recovered
    if getHp() < getMaxHp() then
        sayNext("Hey, your HP is not fully recovered yet. Did you take all the Roger's Apple that I gave you? Are you sure?")
        return
    end
    
    sayNext("How easy is it to consume the item? Simple, right? You can set a #bhotkey#k on the right bottom slot. Haha you didn't know that! right? Oh, and if you are a beginner, HP will automatically recover itself as time goes by. Well it takes time but this is one of the strategies for the beginners.")
    
    sayBoth("Alright! Now that you have learned alot, I will give you a present. This is a must for your travel in Maple World, so thank me! Please use this under emergency cases!")
    
    sayBoth("Okay, this is all I can teach you. I know it's sad but it is time to say good bye. Well take care of yourself and Good luck my friend!\r\n\r\n#fUI/UIWindow2.img/QuestIcon/4/0#\r\n#i2010000# 3 #t2010000#\r\n#i2010009# 3 #t2010009#\r\n\r\n#fUI/UIWindow2.img/QuestIcon/8/0# 10 exp")
    
    -- Give items (TODO: implement inventory check)
    -- giveItem(2010000, 3) -- Apple
    -- giveItem(2010009, 3) -- Green Apple
    
    -- Give EXP
    giveExp(10)
    
    -- Complete the quest
    forceCompleteQuest(1021)
end

