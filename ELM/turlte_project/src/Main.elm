module Main exposing (main)

import Basics exposing (cos, sin)
import Browser
import Html exposing (Html, button, div, textarea)
import Html as H
import Html.Attributes exposing (cols, rows, value)
import Html.Events exposing (onClick, onInput)
import Svg exposing (Svg, circle, line, svg)
import Svg.Attributes as SA


-- LANGUAGE

type Command
    = Forward Int
    | Left Int
    | Right Int
    | Repeat Int (List Command)


-- MODEL

type alias Turtle =
    { x : Float
    , y : Float
    , angle : Float
    }


type alias Segment =
    { x1 : Float
    , y1 : Float
    , x2 : Float
    , y2 : Float
    }


type alias Model =
    { turtle : Turtle
    , segments : List Segment
    , input : String
    }


init : Model
init =
    { turtle = { x = 200, y = 200, angle = 0 }
    , segments = []
    , input = "Repeat 360 [ Forward 1 Left 1 ]"
    }


-- UPDATE

type Msg
    = UpdateInput String
    | Run


update : Msg -> Model -> Model
update msg model =
    case msg of
        UpdateInput str ->
            { model | input = str }

        Run ->
            execute (parse model.input) initialModel


initialModel : Model
initialModel =
    { turtle = { x = 200, y = 200, angle = 0 }
    , segments = []
    , input = ""
    }


-- PARSER

parse : String -> List Command
parse input =
    parseTokens (String.words input)


parseTokens : List String -> List Command
parseTokens tokens =
    case tokens of
        "Forward" :: n :: rest ->
            Forward (toInt n) :: parseTokens rest

        "Left" :: n :: rest ->
            Left (toInt n) :: parseTokens rest

        "Right" :: n :: rest ->
            Right (toInt n) :: parseTokens rest

        "Repeat" :: n :: "[" :: rest ->
            let
                ( block, remaining ) =
                    extractBlock rest
            in
            Repeat (toInt n) (parseTokens block) :: parseTokens remaining

        "]" :: rest ->
            parseTokens rest

        _ ->
            []


extractBlock : List String -> ( List String, List String )
extractBlock tokens =
    extractBlockHelp 1 [] tokens


extractBlockHelp : Int -> List String -> List String -> ( List String, List String )
extractBlockHelp depth acc tokens =
    case tokens of
        "[" :: rest ->
            extractBlockHelp (depth + 1) ("[" :: acc) rest

        "]" :: rest ->
            if depth == 1 then
                ( List.reverse acc, rest )
            else
                extractBlockHelp (depth - 1) ("]" :: acc) rest

        tok :: rest ->
            extractBlockHelp depth (tok :: acc) rest

        [] ->
            ( [], [] )


toInt : String -> Int
toInt s =
    Maybe.withDefault 0 (String.toInt s)


-- EXECUTION

execute : List Command -> Model -> Model
execute cmds model =
    List.foldl executeOne model cmds


executeOne : Command -> Model -> Model
executeOne cmd ({ turtle, segments } as model) =
    case cmd of
        Forward n ->
            forward (toFloat n) model

        Left n ->
            { model | turtle = { turtle | angle = turtle.angle + toFloat n } }

        Right n ->
            { model | turtle = { turtle | angle = turtle.angle - toFloat n } }

        Repeat n cmds ->
            List.foldl (\_ m -> execute cmds m) model (List.range 1 n)


forward : Float -> Model -> Model
forward dist ({ turtle, segments } as model) =
    let
        rad =
            degreesToRadians turtle.angle

        dx =
            dist * cos rad

        dy =
            -dist * sin rad

        x2 =
            turtle.x + dx

        y2 =
            turtle.y + dy

        seg =
            { x1 = turtle.x
            , y1 = turtle.y
            , x2 = x2
            , y2 = y2
            }
    in
    { model
        | turtle = { turtle | x = x2, y = y2 }
        , segments = seg :: segments
    }


degreesToRadians : Float -> Float
degreesToRadians deg =
    deg * pi / 180


-- VIEW

view : Model -> Html Msg
view { turtle, segments, input } =
    div []
        [ textarea
            [ rows 6
            , cols 40
            , value input
            , onInput UpdateInput
            ]
            []
        , button [ onClick Run ] [ H.text "Run" ]
        , svg
            [ SA.width "400"
            , SA.height "400"
            , SA.viewBox "0 0 400 400"
            , SA.style "border:1px solid #ccc"
            ]
            (List.map segmentView segments ++ [ turtleView turtle ])
        ]


segmentView : Segment -> Svg Msg
segmentView s =
    line
        [ SA.x1 (String.fromFloat s.x1)
        , SA.y1 (String.fromFloat s.y1)
        , SA.x2 (String.fromFloat s.x2)
        , SA.y2 (String.fromFloat s.y2)
        , SA.stroke "blue"
        , SA.strokeWidth "2"
        ]
        []


turtleView : Turtle -> Svg Msg
turtleView t =
    circle
        [ SA.cx (String.fromFloat t.x)
        , SA.cy (String.fromFloat t.y)
        , SA.r "3"
        , SA.fill "red"
        ]
        []


-- MAIN

main : Program () Model Msg
main =
    Browser.sandbox
        { init = init
        , update = update
        , view = view
        }
