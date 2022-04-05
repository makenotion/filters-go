package main

import (
    "encoding/json"
    "fmt"
    "os"
)

type Filter struct {
	Predicates []interface{} `json:"-"`// Filter | Predicate
	Operator string `json:"operator"` // and | or
	RawPredicates []json.RawMessage `json:"predicates"`
}

type Predicate struct {
	Column string `json:"column"`
	Test string `json:"test"` // is_empty | is_not_empty
}

func renameQuery(q *Filter, oldCol string, newCol string) *Filter {
	newF := Filter{
		Operator: q.Operator,
	}

	for _, predicate := range q.Predicates {
		fmt.Printf("Renaming %s to %s for predicate %s\n", oldCol, newCol, predicate)
		newF.Predicates = append(newF.Predicates, renameQueryHelper(q, oldCol, newCol))
	}
	return &newF
}

func renameQueryHelper(q interface{}, oldCol string, newCol string) interface{} {
	switch query := q.(type) {
		case *Filter:
			return renameFilter(query, oldCol, newCol)
		case *Predicate:
			return renamePredicate(query, oldCol, newCol)
		default:
			panic("unknown predicate type")
	}
}


func renameFilter(q *Filter, oldCol string, newCol string) *Filter {
	newQ := Filter{
		Operator: q.Operator,
	}
	for _, predicate := range q.Predicates {
		newQ.Predicates = append(newQ.Predicates, renameQueryHelper(predicate, oldCol, newCol))
	}
	return &newQ
}

func renamePredicate(p *Predicate, oldCol string, newCol string) *Predicate {
	if p.Column == oldCol {
		return &Predicate{
			Column: newCol,
			Test: p.Test,
		}
	} else {
		return p
	}
}

func main() {
	str := `{
   "predicates": [
      {
         "column": "Address",
         "test": "is_not_empty"
      },
      {
         "predicates": [
            {
               "column": "Email",
               "test": "is_empty"
            },
            {
               "column": "Phone",
               "test": "is_empty"
            }
         ],
         "operator": "or"
      }
   ],
   "operator": "and"
}`
	var input Filter
	if err := json.Unmarshal([]byte(str), &input); err != nil {
        panic(err)
    }
	output := renameQuery(&input, "Address", "Line 1")

	fmt.Println("Output:")
	enc := json.NewEncoder(os.Stdout)
    enc.Encode(output)
}

// Helpers for de/serializing polymorphic JSON 
func (f *Filter) UnmarshalJSON(b []byte) error {
    type filter Filter
    err := json.Unmarshal(b, (*filter)(f))
    if err != nil {
        return err
    }

    for _, raw := range f.RawPredicates {
        var p map[string]interface{}
        err = json.Unmarshal(raw, &p)
        if err != nil {
			panic(err)
        }

        var i interface{}
		if _, ok := p["column"]; ok {
			i = &Predicate{}
		} else {
			i = &Filter{}
        }
        err = json.Unmarshal(raw, i)
        if err != nil {
            panic(err)
        }
        f.Predicates = append(f.Predicates, i)
    }
    return nil
}

func (f *Filter) MarshalJSON() ([]byte, error) {
    type filter Filter
	if f.Predicates != nil {
		f.RawPredicates = nil
		for _, p := range f.Predicates {
			b, err := json.Marshal(p)
            if err != nil {
                panic(err)
            }
            f.RawPredicates = append(f.RawPredicates, b)
		}
	}
	return json.Marshal((*filter)(f))
}

