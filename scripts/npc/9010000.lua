-- Maple Administrator (9010000)
-- Example NPC script demonstrating all dialogue types

function start()
    -- Get player info
    local name = getPlayerName()
    local level = getPlayerLevel()
    local meso = getPlayerMeso()
    
    -- Greeting with player name (#h is replaced automatically)
    say("Hello, #h#! Welcome to Jinwoo Server.\r\n\r\nI am the #bMaple Administrator#k, here to help you!")
    
    -- Menu selection
    local choice = askMenu("What would you like to do?\r\n\r\n#L0##bGet some starter items#l\r\n#L1##bTeleport to Henesys#l\r\n#L2##bCheck my stats#l\r\n#L3##bTest Yes/No dialogue#l\r\n#L4##bTest number input#l")
    
    if choice == 0 then
        handleStarterItems()
    elseif choice == 1 then
        handleTeleport()
    elseif choice == 2 then
        handleStats(name, level, meso)
    elseif choice == 3 then
        handleYesNo()
    elseif choice == 4 then
        handleNumber()
    end
end

function handleStarterItems()
    local accept = askAcceptDecline("Would you like to receive some starter items?\r\n\r\n#fUI/UIWindow.img/QuestIcon/4/0#\r\n#v2000000# #t2000000# x 100\r\n#v2000003# #t2000003# x 50")
    
    if accept then
        -- Give items (will log for now until inventory system is implemented)
        local success1 = giveItem(2000000, 100)  -- Red Potion x100
        local success2 = giveItem(2000003, 50)   -- Orange Potion x50
        
        if success1 and success2 then
            say("Here you go! Good luck on your adventures!")
        else
            say("I'm sorry, but your inventory seems to be full.")
        end
    else
        say("No problem! Come back anytime you need help.")
    end
end

function handleTeleport()
    local confirm = askYesNo("Would you like to be teleported to #bHenesys#k?")
    
    if confirm then
        say("Alright, off you go!")
        warp(100000000, 0)  -- Henesys
    else
        say("Maybe next time then!")
    end
end

function handleStats(name, level, meso)
    local msg = "Here are your current stats:\r\n\r\n"
    msg = msg .. "#eName:#n " .. name .. "\r\n"
    msg = msg .. "#eLevel:#n " .. level .. "\r\n"
    msg = msg .. "#eMeso:#n " .. meso .. "\r\n"
    msg = msg .. "#eJob:#n " .. getPlayerJob() .. "\r\n"
    msg = msg .. "#eMap:#n " .. getPlayerMap()
    
    say(msg)
end

function handleYesNo()
    local answer = askYesNo("This is a #bYes/No#k question.\r\n\r\nDo you like this server?")
    
    if answer then
        say("Thank you! We're glad you're enjoying it!")
    else
        say("Oh no! Please let us know how we can improve.")
    end
end

function handleNumber()
    local num = askNumber("Enter a number between 1 and 100:", 50, 1, 100)
    
    say("You entered: #b" .. num .. "#k\r\n\r\nThat's a great number!")
end

