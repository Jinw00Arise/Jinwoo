--[[
    Quest Script: q1021s
    Roger's Apple (1021 - start)
    NPC: Roger (2000)
]]

function start()
    local gender = getGender()
    local greeting = gender == 0 and "Man" or "Miss"
    
    sayNext("Hey, " .. greeting .. "~ What's up? Haha! I am Roger who teaches you new travellers with lots of information.")
    
    sayBoth("You are asking who made me do this? Ahahahaha! Myself! I wanted to do this and just be kind to you new travellers.")
    
    if not askAcceptDecline("So..... Let me just do this for fun! Abaracadabra~!") then
        return
    end
    
    -- Damage the player to demonstrate HP system
    setHp(25)
    
    sayNext("Surprised? If HP becomes 0, then you are in trouble. Now, I will give you  #rRoger's Apple#k. Please take it. You will feel stronger. Open the item window and double click to consume. Hey, It's very simple to open the item window. Just press #bI#k on your keyboard.")
    
    sayBoth("Please take all Roger's Apples that I gave you. You will be able to see the HP bar increasing right away. Please talk to me again when you recover your HP 100%.")
    
    -- Give Roger's Apple
    if not hasItem(2010007, 1) then
        if not giveItem(2010007, 1) then
            sayNext("Please check if your inventory is full or not.")
            return
        end
    end
    
    -- Start the quest
    forceStartQuest(1021)
    
    -- Show tutorial UI
    avatarOriented("UI/tutorial.img/28")
end

