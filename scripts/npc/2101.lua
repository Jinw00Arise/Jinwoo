-- Lith Harbor NPC (2101) 
-- Simple quest/guide NPC example

function start()
    say("Welcome to #bLith Harbor#k, #h#!\r\n\r\nThis is where all new adventurers begin their journey in Victoria Island.")
    
    local choice = askMenu("How can I help you?\r\n\r\n#L0##bTell me about this place#l\r\n#L1##bI want to travel somewhere#l\r\n#L2##bGive me some tips#l")
    
    if choice == 0 then
        aboutPlace()
    elseif choice == 1 then
        travel()
    elseif choice == 2 then
        tips()
    end
end

function aboutPlace()
    say("Lith Harbor is the main port town of Victoria Island. Ships come and go from here to various destinations.\r\n\r\nTo the east lies #bHenesys#k, the town of bowmen.\r\nTo the south is #bKerning City#k, home of the thieves.")
end

function travel()
    say("I'm sorry, but the ship service is currently unavailable.\r\n\r\nPlease use the portals to travel to nearby towns, or ask the #bMaple Administrator#k for help.")
end

function tips()
    local level = getPlayerLevel()
    
    if level < 10 then
        say("As a new adventurer, I recommend:\r\n\r\n#e1.#n Train on Snails and Blue Snails nearby\r\n#e2.#n Complete beginner quests for rewards\r\n#e3.#n Save your mesos for potions\r\n#e4.#n Reach level 10 to choose your job!")
    elseif level < 30 then
        say("You're making good progress! Here are some tips:\r\n\r\n#e1.#n Try hunting in the Henesys Hunting Ground\r\n#e2.#n Join a party for faster leveling\r\n#e3.#n Don't forget to allocate your AP and SP")
    else
        say("You're an experienced adventurer! Keep up the great work.\r\n\r\nConsider exploring:\r\n#e1.#n Kerning PQ (level 21-30)\r\n#e2.#n Orbis Tower\r\n#e3.#n Ludibrium")
    end
end

