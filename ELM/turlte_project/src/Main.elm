module Main exposing (main)

import Basics exposing (cos, sin)
import Browser
import Html exposing (Html, button, div, textarea)
import Html as H
import Html.Attributes exposing (checked, cols, rows, type_, value)
import Html.Attributes as HA
import Html.Events exposing (onClick, onInput, onCheck)
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
    , penDown : Bool
    , color : String
    }


type alias Segment =
    { x1 : Float
    , y1 : Float
    , x2 : Float
    , y2 : Float
    , color : String
    }


type alias DrawingState =
    { turtle : Turtle
    , segments : List Segment
    }


type alias Model =
    { current : DrawingState
    , preview : Maybe DrawingState
    , undoStack : List DrawingState
    , redoStack : List DrawingState
    , input : String
    }



-- INIT


init : Model
init =
    { current =
        { turtle =
            { x = 200
            , y = 200
            , angle = 0
            , penDown = True
            , color = "#0000ff"
            }
        , segments = []
        }
    , preview = Nothing
    , undoStack = []
    , redoStack = []
    , input = "Repeat 360 [ Forward 1 Left 1 ]"
    }



-- UPDATE


type Msg
    = UpdateInput String
    | Run
    | Undo
    | Redo
    | TogglePen Bool
    | UpdateColor String


update : Msg -> Model -> Model
update msg model =
    case msg of
        UpdateInput str ->
            let
                cmds =
                    parse str

                newPreview =
                    case cmds of
                        [] ->
                            Nothing

                        _ ->
                            Just (execute cmds model.current)
            in
            { model
                | input = str
                , preview = newPreview
            }

        Run ->
            let
                cmds =
                    parse model.input

                newCurrent =
                    execute cmds model.current
            in
            { model
                | current = newCurrent
                , undoStack = model.current :: model.undoStack
                , redoStack = []
                , preview = Nothing
            }

        Undo ->
            case model.undoStack of
                prev :: rest ->
                    { model
                        | current = prev
                        , undoStack = rest
                        , redoStack = model.current :: model.redoStack
                        , preview = Nothing
                    }

                [] ->
                    model

        Redo ->
            case model.redoStack of
                next :: rest ->
                    { model
                        | current = next
                        , redoStack = rest
                        , undoStack = model.current :: model.undoStack
                        , preview = Nothing
                    }

                [] ->
                    model

        TogglePen value ->
            let
                oldCurrent =
                    model.current

                oldTurtle =
                    oldCurrent.turtle

                newTurtle =
                    { oldTurtle | penDown = value }

                newCurrent =
                    { oldCurrent | turtle = newTurtle }

                cmds =
                    parse model.input

                newPreview =
                    case cmds of
                        [] ->
                            Nothing

                        _ ->
                            Just (execute cmds newCurrent)
            in
            { model
                | current = newCurrent
                , preview = newPreview
            }

        UpdateColor color ->
            let
                oldCurrent =
                    model.current

                oldTurtle =
                    oldCurrent.turtle

                newTurtle =
                    { oldTurtle | color = color }

                newCurrent =
                    { oldCurrent | turtle = newTurtle }

                cmds =
                    parse model.input

                newPreview =
                    case cmds of
                        [] ->
                            Nothing

                        _ ->
                            Just (execute cmds newCurrent)
            in
            { model
                | current = newCurrent
                , preview = newPreview
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


execute : List Command -> DrawingState -> DrawingState
execute cmds state =
    List.foldl executeOne state cmds


executeOne : Command -> DrawingState -> DrawingState
executeOne cmd state =
    case cmd of
        Forward n ->
            forward (toFloat n) state

        Left n ->
            let
                turtle =
                    state.turtle

                newTurtle =
                    { turtle | angle = turtle.angle + toFloat n }
            in
            { state | turtle = newTurtle }

        Right n ->
            let
                turtle =
                    state.turtle

                newTurtle =
                    { turtle | angle = turtle.angle - toFloat n }
            in
            { state | turtle = newTurtle }

        Repeat n cmds ->
            List.foldl (\_ s -> execute cmds s) state (List.range 1 n)


forward : Float -> DrawingState -> DrawingState
forward dist state =
    let
        turtle =
            state.turtle

        rad =
            degreesToRadians turtle.angle

        dx =
            dist * cos rad

        dy =
            -dist * sin rad

        newX =
            turtle.x + dx

        newY =
            turtle.y + dy

        newTurtle =
            { turtle | x = newX, y = newY }
    in
    if turtle.penDown then
        let
            seg =
                { x1 = turtle.x
                , y1 = turtle.y
                , x2 = newX
                , y2 = newY
                , color = turtle.color
                }
        in
        { state
            | turtle = newTurtle
            , segments = seg :: state.segments
        }

    else
        { state | turtle = newTurtle }


degreesToRadians : Float -> Float
degreesToRadians deg =
    deg * pi / 180



-- VIEW


view : Model -> Html Msg
view model =
    let
        state =
            model.current
    in
    div
        [ HA.style "display" "flex"
        , HA.style "flex-direction" "column"
        , HA.style "align-items" "center"
        , HA.style "margin-top" "20px"
        , HA.style "gap" "10px"
        ]
        [ svg
            [ SA.width "400"
            , SA.height "400"
            , SA.viewBox "0 0 400 400"
            , SA.style "border:1px solid #ccc"
            ]
            (List.map segmentView state.segments
                ++ previewView model.preview
                ++ [ turtleView state.turtle ]
            )
        , div []
            [ button [ onClick Undo ] [ H.text "←" ]
            , button [ onClick Redo ] [ H.text "→" ]
            ]
        , div []
            [ H.label []
                [ H.text "Pen down "
                , H.input
                    [ type_ "checkbox"
                    , checked state.turtle.penDown
                    , onCheck TogglePen
                    ]
                    []
                ]
            ]
        , div []
            [ H.text "Color (hex) "
            , H.input
                [ type_ "text"
                , value state.turtle.color
                , onInput UpdateColor
                ]
                []
            ]
        , textarea
            [ rows 6
            , cols 40
            , value model.input
            , onInput UpdateInput
            ]
            []
        , button [ onClick Run ] [ H.text "Run" ]
        ]


segmentView : Segment -> Svg Msg
segmentView s =
    line
        [ SA.x1 (String.fromFloat s.x1)
        , SA.y1 (String.fromFloat s.y1)
        , SA.x2 (String.fromFloat s.x2)
        , SA.y2 (String.fromFloat s.y2)
        , SA.stroke s.color
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


previewView : Maybe DrawingState -> List (Svg Msg)
previewView maybeState =
    case maybeState of
        Nothing ->
            []

        Just state ->
            List.map previewSegmentView state.segments


previewSegmentView : Segment -> Svg Msg
previewSegmentView s =
    line
        [ SA.x1 (String.fromFloat s.x1)
        , SA.y1 (String.fromFloat s.y1)
        , SA.x2 (String.fromFloat s.x2)
        , SA.y2 (String.fromFloat s.y2)
        , SA.stroke s.color
        , SA.strokeOpacity "0.3"
        , SA.strokeWidth "2"
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
