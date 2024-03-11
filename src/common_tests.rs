#[cfg(test)]
pub mod tests {
    use crate::TantivySession;
    use super::*;

    #[test]
    fn test_get_words() {

        // Test case 1
        let query1 = "Hello World";

        let words1 = TantivySession::get_words(query1);
        assert_eq!(words1, vec!["hello", "world"]);

        // Test case 2
        let query2 = "  Rust    is   awesome  ";
        let words2 = TantivySession::get_words(query2);
        assert_eq!(words2, vec!["rust", "is", "awesome"]);

        // Test case 3: Empty string
        let query3 = "";
        let words3 = TantivySession::get_words(query3);
        assert_eq!(words3, Vec::<String>::new());
    }
}