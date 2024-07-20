CREATE TABLE "achievement_names" (
    "id" integer NOT NULL,
    "name" character varying(128) NOT NULL,
    PRIMARY KEY ("id")
);

INSERT INTO "achievement_names" ("id", "name") VALUES
    (-1, 'WvW Rank'),
    (-2, 'Playtime'),
    (283, 'Kills'),
    (306, 'Supply Spend'),
    (285, 'Dolly Escort'),
    (288, 'Dolly Kill'),
    (303, 'Objective Capture'),
    (319, 'Objective Defend'),
    (291, 'Camp Capture'),
    (310, 'Camp Defend'),
    (297, 'Tower Capture'),
    (322, 'Tower Defend'),
    (300, 'Keep Capture'),
    (316, 'Keep Defend'),
    (294, 'Castle Capture'),
    (313, 'Castle Defend');