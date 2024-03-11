use itertools::Itertools;
use regex::Regex;
use crate::{debug};
use crate::make_internal_json_error;
use crate::ErrorKinds;
use crate::InternalCallResult;
use crate::TantivySession;

extern crate serde;
extern crate serde_derive;
extern crate serde_json;

use tantivy::query::{BooleanQuery, BoostQuery, DisjunctionMaxQuery, FuzzyTermQuery, Occur, PhrasePrefixQuery, PhraseQuery, Query, QueryParserError, RegexQuery, TermQuery, TermSetQuery};
use tantivy::query::QueryParser;
use tantivy::schema::{Field, IndexRecordOption, Term};
use tantivy::TantivyError;

impl TantivySession {
    pub fn handle_query_parser(
        &mut self,
        method: &str,
        params: serde_json::Value,
    ) -> InternalCallResult<u32> {
        let m = match params.as_object() {
            Some(m) => m,
            None => {
                return make_internal_json_error::<u32>(ErrorKinds::BadParams(
                    "invalid parameters pass to query_parser add_text".to_string(),
                ));
            }
        };
        debug!("QueryParser");
        if method == "for_index" {
            let mut v_out: Vec<Field> = Vec::<Field>::new();
            debug!("QueryParser aquired");
            let schema = match self.schema.as_ref() {
                Some(s) => s,
                None => {
                    return make_internal_json_error(ErrorKinds::BadInitialization(
                        "schema not available during for_index".to_string(),
                    ));
                }
            };
            let request_fields = m
                .get("fields")
                .and_then(|f| f.as_array())
                .ok_or_else(|| ErrorKinds::BadParams("fields not present".to_string()))?;
            for v in request_fields {
                let v_str = v.as_str().unwrap_or_default();
                if let Ok(f) = schema.get_field(v_str) {
                    v_out.append(vec![f].as_mut())
                }
            }
            let tm = match &self.tokenizer_manager {
                Some(tm) => tm,
                None => {
                    return make_internal_json_error(ErrorKinds::BadInitialization(
                        "token manager not available on session".to_string(),
                    ));
                }
            };
            self.query_parser = Some(Box::new(QueryParser::new(
                schema.clone(),
                v_out,
                tm.clone(),
            )));
            return Ok(0);
        }
        if method == "parse_query" {
            let qp = match &self.query_parser {
                Some(qp) => qp,
                None => {
                    return make_internal_json_error::<u32>(ErrorKinds::NotExist(
                        "index is None".to_string(),
                    ));
                }
            };
            let query = match m.get("query") {
                Some(q) => match q.as_str() {
                    Some(s) => s,
                    None => {
                        return make_internal_json_error::<u32>(ErrorKinds::BadParams(
                            "query parameter must be a string".to_string(),
                        ));
                    }
                },
                None => {
                    return make_internal_json_error::<u32>(ErrorKinds::BadParams(
                        "parameter 'query' missing".to_string(),
                    ));
                }
            };
            // self.dyn_q = match qp.parse_query(regex::escape(query).as_str()) {
            //     Ok(qp) => Some(qp),
            //     Err(e) => {
            //         return make_internal_json_error::<u32>(ErrorKinds::BadParams(format!(
            //             "query parser error : {e}"
            //         )));
            //     }
            // };

            let (dyn_q, err) = qp.parse_query_lenient(regex::escape(query).as_str());
            self.dyn_q = Option::from(dyn_q);
            return Ok(0);
        }
        if method == "prepare_query" {
            let qp = match &self.query_parser {
                Some(qp) => qp,
                None => {
                    return make_internal_json_error::<u32>(ErrorKinds::NotExist(
                        "index is None".to_string(),
                    ));
                }
            };
            let query = match m.get("query") {
                Some(q) => match q.as_str() {
                    Some(s) => s,
                    None => {
                        return make_internal_json_error::<u32>(ErrorKinds::BadParams(
                            "query parameter must be a string".to_string(),
                        ));
                    }
                },
                None => {
                    return make_internal_json_error::<u32>(ErrorKinds::BadParams(
                        "parameter 'query' missing".to_string(),
                    ));
                }
            };
            let space_id = match m.get("space_id") {
                Some(q) => match q.as_str() {
                    Some(s) => s,
                    None => {
                        return make_internal_json_error::<u32>(ErrorKinds::BadParams(
                            "space_id parameter must be a string".to_string(),
                        ));
                    }
                },
                None => {
                    return make_internal_json_error::<u32>(ErrorKinds::BadParams(
                        "parameter 'space_id' missing".to_string(),
                    ));
                }
            };
            let regex = Regex::new("[\\+\\^`:\\{\\}\\\"\\[\\]\\(\\~\\!\\*\\(\\)]|\\\\\\\\").unwrap();
            let (dyn_q, err) = qp.parse_query_lenient(regex.replace_all("+^`:{}\"[](~!\\\\\\*()", "").to_string().as_str());
            self.dyn_q = match self.make_queries(query, space_id, dyn_q) {
                Ok(qp) => Some(qp),
                Err(e) => {
                    return make_internal_json_error::<u32>(ErrorKinds::BadParams(format!(
                        "query parser error : {e}"
                    )));
                }
            };
            return Ok(0);
        }
        if method == "parse_fuzzy_query" {
            let schema = match self.schema.as_ref() {
                Some(s) => s,
                None => {
                    return make_internal_json_error(ErrorKinds::BadInitialization(
                        "schema not available during for_index".to_string(),
                    ));
                }
            };
            let request_field = m
                .get("field")
                .and_then(|f| f.as_array())
                .ok_or_else(|| ErrorKinds::BadParams("field not present".to_string()))?;
            if request_field.len() != 1 {
                return make_internal_json_error(ErrorKinds::BadInitialization(
                    "Requesting more than one field in fuzzy query disallowed".to_string(),
                ));
            }
            let fuzzy_term = m
                .get("term")
                .and_then(|f| f.as_array())
                .ok_or_else(|| ErrorKinds::BadParams("term not present".to_string()))?;
            if fuzzy_term.len() != 1 {
                return make_internal_json_error(ErrorKinds::BadInitialization(
                    "Requesting more than one term in fuzzy query disallowed".to_string(),
                ));
            }

            let field = &request_field[0];
            let f_str = match field.as_str() {
                Some(s) => s,
                None => {
                    return make_internal_json_error(ErrorKinds::BadInitialization(
                        "Field requested is not present".to_string(),
                    ));
                }
            };
            if let Ok(f) = schema.get_field(f_str) {
                let f_term = fuzzy_term[0].as_str().ok_or(ErrorKinds::BadInitialization(
                    "Failed to parse fuzzy term".to_string(),
                ))?;
                let t = Term::from_field_text(f, f_term);
                let q = FuzzyTermQuery::new(t, 1, true);
                self.fuzzy_q = Some(Box::new(q));
            }
            return Ok(0);
        }
        let e = ErrorKinds::BadParams(format!("Unknown method {method}"));
        Err(e)
    }

    // PhraseQuery::new(Self::get_terms(query, title));
    // PhraseQuery::new(Self::get_terms(query, body));
    // PhrasePrefixQuery::new(
    //     Self::get_terms(query, title)
    //         .iter()
    //         .take(5)
    //         .cloned()
    //         .collect()
    // );
    //PhrasePrefixQuery::new(Self::get_terms(query, body).iter().take(5).collect());
    // let terms_title = Self::get_terms(query, title)
    //     .into_iter()
    //     .map(|term|
    //         Box::new(TermQuery::new(term, IndexRecordOption::WithFreqsAndPositions)) as Box<dyn Query>
    //     )
    //     .collect();
    // let terms_body = Self::get_terms(query, body)
    //     .into_iter()
    //     .map(|term|
    //         Box::new(TermQuery::new(term, IndexRecordOption::WithFreqsAndPositions)) as Box<dyn Query>
    //     )
    //     .collect();
    // DisjunctionMaxQuery::new(terms_title);
    //todo in the end

    fn make_queries(&mut self, query: &str, space_id: &str, dyn_q: Box<dyn Query>) -> Result<Box<dyn Query>, TantivyError> {
        let query = query.to_lowercase().trim().to_owned();
        let field_id = 0;
        let field_space = 1;
        let field_title = 2;
        let field_title_no_terms = 3;
        let field_text = 4;
        let field_text_no_terms = 5;

        let title_regex = match Self::prepare_regex_query(field_title_no_terms, &query) {
            Ok(rq) => Some(rq),
            Err(e) => { return Err(e); }
        };

        let text_regex = match Self::prepare_regex_query(field_text_no_terms, &query) {
            Ok(rq) => Some(rq),
            Err(e) => { return Err(e); }
        };

        let mut queries: Vec<(Occur, Box<dyn Query>)> = vec![];
        if !space_id.is_empty() {
            queries.push(
                (Occur::Must, Box::new(TermQuery::new(
                    Term::from_field_text(Field::from_field_id(field_space), space_id),
                    IndexRecordOption::Basic)
                ))
            )
        }
        //debug!("### field_title {:?}", TermSetQuery::new(Self::get_terms(&query, field_title)));
        //debug!("### field_text {:?}", TermSetQuery::new(Self::get_terms(&query, field_text)));
        debug!("### title_regex {:?}", title_regex);
        debug!("### text_regex {:?}", text_regex);
        queries.push((Occur::Must, Box::new(DisjunctionMaxQuery::new(vec![
            Box::new(BoostQuery::new(Box::new(TermQuery::new(
                Term::from_field_text(Field::from_field_id(field_id), &query),
                IndexRecordOption::Basic)
            ), 30.0)),
            dyn_q,
            Box::new(title_regex.unwrap()),
            Box::new(text_regex.unwrap()),
        ]))));

        return Ok(Box::new(BooleanQuery::new(queries)));
    }

    fn prepare_regex_query(field: u32, query: &str) -> tantivy::Result<RegexQuery> {
        return RegexQuery::from_pattern(
            Self::get_words(&query)
                .into_iter()
                .enumerate()
                .map(|(i, term)| {
                    if i == 0 {
                        format!("{}{}{}", ".*", regex::escape(term.as_str()), ".*")
                    } else {
                        format!("{}{}", regex::escape(term.as_str()), ".*")
                    }
                })
                .collect::<Vec<_>>()
                .join("")
                .as_str()
            ,
            Field::from_field_id(field),
        );
    }

    fn get_terms(query: &str, field_id: u32) -> Vec<Term> {
        Self::get_words(query)
            .into_iter()
            .map(|term| Term::from_field_text(Field::from_field_id(field_id), term.as_str())) //todo field id?
            .collect()
    }

    pub fn get_words(query: &str) -> Vec<String> {
        query.split_whitespace()
            .filter(|e| !e.is_empty())
            .map(|e| e.to_owned())
            .collect_vec()
    }
}