package main

import (
  "bufio"
  "fmt"
  "os"
  "regexp"
  "strings"
)

// Interfaces

// Rule defined single openflow Rule in the table
type Rule struct {
  id int
  tableId string
  matchAndAct string
  nextHop string
  rawRule string
}

// Table represents OpenFlow table with rule
type Table struct {
  tableId string
  rules []*Rule
}

// FlowsDumpReader responsible for reading text file with dump of OpenFlow tables
type FlowsDumpReader struct {
  file string
}

func (fdr *FlowsDumpReader) read() map[string]*Table {
  var tables = make(map[string]*Table)

  // read file
  reader, err := os.Open(fdr.file)
  if err != nil {
    fmt.Println("Failed to open file:", fdr.file)
  }
  defer reader.Close()

  bufReader := bufio.NewReader(reader)
  var line string

  tableRegex, _ := regexp.Compile("table=([0-9]+)")
  nextTableRegex, _ := regexp.Compile("resubmit\\(,([0-9]+)\\)")
  matchAndActRegex, _ := regexp.Compile("priority=[0-9]+,?(.*)")
  sendToRegex, _ := regexp.Compile("actions=([A-Z]+:?([0-9]+)?)")
  var ruleId int
  for {
    line, err = bufReader.ReadString('\n')
    if err != nil {
      break
    }
    cleanLine := strings.TrimSpace(line)
    if len(cleanLine) > 0 {
       tableIdMatched := tableRegex.FindStringSubmatch(cleanLine)
       matchAndActMatched := matchAndActRegex.FindStringSubmatch(cleanLine)
       sendToMatched := sendToRegex.FindStringSubmatch(matchAndActMatched[1])
       nextTableMatched := nextTableRegex.FindStringSubmatch(matchAndActMatched[1])

       rule := &Rule {
         id: ruleId,
         tableId: tableIdMatched[1],
         rawRule: cleanLine,
         matchAndAct: matchAndActMatched[1],
       }
       table, tableExists := tables[tableIdMatched[1]]
       if !tableExists {
         table = &Table {
           tableId: tableIdMatched[1],
         }
         tables[tableIdMatched[1]] = table
       }
       table.rules = append(table.rules, rule)

       if len(sendToMatched) > 1 {
         outTableName := sendToMatched[1]
         outTableName = strings.Replace(outTableName, ":", "_", -1)

         sendToTable, tableExists := tables[outTableName]
        rule.nextHop = outTableName
        if !tableExists {
          sendToTable = &Table {
            tableId: outTableName,
          }
          tables[outTableName] = sendToTable
        }

       }
      if len(nextTableMatched) > 1 {
        nextTable, tableExists := tables[nextTableMatched[1]]
        rule.nextHop = nextTableMatched[1]
        if !tableExists {
          nextTable = &Table {
            tableId: nextTableMatched[1],
          }
          tables[nextTableMatched[1]] = nextTable
        }
      }

      ruleId++

    }
  }

  return tables
}

//PlantUmlRenderer responsible for generation representation of the Open Flow table in PlanUML notation
type PlantUmlRenderer struct {
  file   string
  tables map[string]*Table
}

func (pur *PlantUmlRenderer) render() {
   outFile, err := os.Create(pur.file)
   defer outFile.Close()

   w := bufio.NewWriter(outFile)
   if err != nil {}

   fmt.Fprintln(w,"@startuml")
   fmt.Fprintln(w,"left to right direction")
   fmt.Fprintln(w,"top to bottom direction")

   for _, table := range pur.tables {
     const TABLE = "Table"
     const RULE = "Rule"

     fmt.Fprintln(w, "rectangle", getObjName(TABLE, table.tableId), "{")
     for _, rule := range table.rules {
          fmt.Fprintln(w,"folder", getObjName(RULE, rule.id), "[")
          fmt.Fprintln(w,"name", getObjName(RULE, rule.id))
          fmt.Fprintln(w, rule.matchAndAct)
          fmt.Fprintln(w,"]")
        }
        fmt.Fprintln(w,"}")
        fmt.Fprintln(w,"")
       for _, rule := range table.rules {
         if len(rule.nextHop) > 0 {
           fmt.Fprintln(w, getObjName(RULE, rule.id), "------>", getObjName(TABLE, rule.nextHop))
         }
       }

   }
   fmt.Fprintln(w,"@enduml")
   w.Flush()

}

// Functions
func main() {
  args := os.Args[1:]

  reader := FlowsDumpReader{
    file: args[0],
  }
  tables := reader.read()

  renderer := PlantUmlRenderer{
    file:   "./out.puml",
    tables: tables,
  }
  renderer.render()

}

func getObjName(prefix string, postfix interface{}) string {
 return fmt.Sprintf("%s_%v", prefix, postfix)
}