Attribute VB_Name = "ResumeMatcher"
Option Explicit

Private Function Q(ByVal s As String) As String
    Q = """" & s & """"
End Function

Public Sub SelectJDFile()
    Dim fd As FileDialog
    Set fd = Application.FileDialog(msoFileDialogFilePicker)
    fd.Title = "Select Job Description File"
    fd.AllowMultiSelect = False
    fd.Filters.Clear
    fd.Filters.Add "Supported", "*.txt;*.text;*.md;*.pdf;*.docx;*.rtf"
    If fd.Show = -1 Then
        Sheets("Inputs").Range("B1").Value = fd.SelectedItems(1)
    End If
End Sub

Public Sub Auto_Open()
    CreateButtons
End Sub

Public Sub SelectResumesFolder()
    Dim fd As FileDialog
    Set fd = Application.FileDialog(msoFileDialogFolderPicker)
    fd.Title = "Select Resumes Folder"
    If fd.Show = -1 Then
        Sheets("Inputs").Range("B2").Value = fd.SelectedItems(1)
    End If
End Sub

Public Sub RunMatcher()
    Dim wsInputs As Worksheet
    Dim wsResults As Worksheet
    Dim matcherExe As String
    Dim projectRoot As String
    Dim scriptPath As String
    Dim wbPath As String
    Dim outCsv As String
    Dim cmd As String
    Dim wsh As Object
    Dim execObj As Object
    Dim exitCode As Long
    Dim jdPath As String
    Dim resumesPath As String
    Dim stdErr As String
    Dim stdOut As String

    Set wsInputs = Sheets("Inputs")
    Set wsResults = Sheets("Results")

    jdPath = Trim(wsInputs.Range("B1").Value)
    resumesPath = Trim(wsInputs.Range("B2").Value)

    matcherExe = Trim(wsInputs.Range("B7").Value)

    projectRoot = Trim(wsInputs.Range("B8").Value)
    If projectRoot = "" Then
        projectRoot = ThisWorkbook.Path
    End If

    If projectRoot = "" Then
        MsgBox "Project root not set. Please fill Inputs!B8 or save workbook in project folder.", vbExclamation
        Exit Sub
    End If

    scriptPath = projectRoot & "\bin\resume_matcher.exe"
    wbPath = ThisWorkbook.FullName
    outCsv = Trim(wsInputs.Range("B4").Value)
    If outCsv = "" Then
        outCsv = projectRoot & "\outputs\results.csv"
    End If

    wsInputs.Range("B5").Value = Now
    wsInputs.Range("B6").Value = "Running..."

    ThisWorkbook.Save

    If matcherExe = "" Then
        matcherExe = scriptPath
    End If

    If Dir(matcherExe) = "" Then
        wsInputs.Range("B6").Value = "Failed (exe missing)"
        MsgBox "resume_matcher.exe not found at: " & matcherExe, vbExclamation
        Exit Sub
    End If
    If jdPath = "" Or Dir(jdPath) = "" Then
        wsInputs.Range("B6").Value = "Failed (JD missing)"
        MsgBox "Job description file not found: " & jdPath, vbExclamation
        Exit Sub
    End If
    If resumesPath = "" Or Dir(resumesPath, vbDirectory) = "" Then
        wsInputs.Range("B6").Value = "Failed (Resumes folder missing)"
        MsgBox "Resumes folder not found: " & resumesPath, vbExclamation
        Exit Sub
    End If

    cmd = Q(matcherExe) & " --workbook " & Q(wbPath)

    Set wsh = CreateObject("WScript.Shell")
    On Error GoTo ExecFailed
    Set execObj = wsh.Exec(cmd)
    On Error GoTo 0

    Do While execObj.Status = 0
        DoEvents
    Loop
    exitCode = execObj.ExitCode
    stdErr = execObj.StdErr.ReadAll
    stdOut = execObj.StdOut.ReadAll

    If exitCode <> 0 Then
        wsInputs.Range("B6").Value = "Failed (code " & exitCode & ")"
        MsgBox "Python failed." & vbCrLf & _
               "Command: " & cmd & vbCrLf & _
               "StdErr: " & vbCrLf & stdErr & vbCrLf & _
               "StdOut: " & vbCrLf & stdOut, vbExclamation
        Exit Sub
    End If

    wsInputs.Range("B6").Value = "Completed"
    ImportResultsCSV wsResults, outCsv
    MsgBox "Done. Results are in the Results sheet.", vbInformation
    Exit Sub

ExecFailed:
    wsInputs.Range("B6").Value = "Failed (exec)"
    MsgBox "Failed to launch matcher." & vbCrLf & _
           "Command: " & cmd & vbCrLf & _
           "Error: " & Err.Description, vbExclamation
End Sub

Public Sub CreateButtons()
    Dim ws As Worksheet
    Dim btn As Button
    Dim leftPos As Double
    Dim topPos As Double
    Dim btnWidth As Double
    Dim btnHeight As Double

    Set ws = Sheets("Inputs")

    leftPos = ws.Range("D1").Left
    topPos = ws.Range("D1").Top
    btnWidth = 140
    btnHeight = 28

    On Error Resume Next
    ws.Buttons.Delete
    On Error GoTo 0

    Set btn = ws.Buttons.Add(leftPos, topPos, btnWidth, btnHeight)
    btn.Caption = "Select JD"
    btn.OnAction = "SelectJDFile"

    Set btn = ws.Buttons.Add(leftPos, topPos + 36, btnWidth, btnHeight)
    btn.Caption = "Select Resumes"
    btn.OnAction = "SelectResumesFolder"

    Set btn = ws.Buttons.Add(leftPos, topPos + 72, btnWidth, btnHeight)
    btn.Caption = "Run Matcher"
    btn.OnAction = "RunMatcher"
End Sub

Private Sub ImportResultsCSV(ByVal ws As Worksheet, ByVal csvPath As String)
    ws.Cells.Clear

    With ws.QueryTables.Add(Connection:="TEXT;" & csvPath, Destination:=ws.Range("A1"))
        .TextFileParseType = xlDelimited
        .TextFileCommaDelimiter = True
        .TextFileTextQualifier = xlTextQualifierDoubleQuote
        .Refresh BackgroundQuery:=False
        .Delete
    End With
End Sub
