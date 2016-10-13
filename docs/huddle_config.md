Annotated Huddle Configuration File
===================================

TODO

The following annotated huddle configuration file describes the configurable points used to schedule huddles in Intervention Engine.  Note that the JSON specification does *not* allow for inline comments, so this annotated version is technically invalid JSON.  Please ensure any huddle configuration file you use does *not* contain any comments.

```js
{
  /* Name is a human readable name for the huddle */
  "name": "Example Huddle",
  /* LeaderID refers to the ID of the practitioner or group that leads the huddle.  Each unique huddle team needs
     its own leaderID, but other than that it is not currently used (and does not need to correspond to a real ID).
     For a single-huddle group implementation, leave as 1. */
  "leaderID": "1",
  /* Days indicates the days of the week on which the huddle meets.  Separate more than one day using a comma
     (e.g., [1, 3, 5]).  The example below indicates the group meets on Mondays.  Use the following key to determine
     the integers to use: Sunday: 0, Monday: 1, Tuesday: 2, Wednesday: 3, Thursday: 4, Friday: 5, Saturday: 6 */
  "days": [1],
  /* LookAhead indicates the number of huddles to schedule forward.  In this example, whenever the scheduling algorithm
     is run, it will schedule up to 14 huddles in advance. */
  "lookAhead": 14,
  /* RiskConfig contains the configuration for scheduling huddles based on risk scores.  This section can be removed if
     you do not wish to use risk scores as a factor for scheduling huddles. */
  "riskConfig": {
    /* RiskMethod needs to correspond to the method code used in the FHIR RiskAssessment resources.  It will be used
       to lookup the risk scores of patients. */
    "riskMethod": {"system": "http://interventionengine.org/risk-assessments", "code": "ExampleRisk"},
    /* FrequencyConfigs is a list of configurations indicating the frequency at which patients should be scheduled,
       based on their risk score.  Usually, the higher the score, the more frequently they should be scheduled.
       This example uses a fiction risk scoring algorithm that range from 0 to 20. */
    "frequencyConfigs": [
      /* This config indicates that patients with a risk score of 18-20 should be discussed about once a week. */
      {
        /* MinScore indicates the low (inclusive) value of the risk score indicating his frequency configuration. */
        "minScore": 18,
        /* MaxScore indicates the high (inclusive) value of the risk score indicating this frequency configuration. */
        "maxScore": 20,
        /* MinDaysBetweenHuddles indicates the smallest number of days that should pass between huddles in which this
           patient is discussed.  In other words, discuss this patient no more frequently than every 5 days. */
        "minDaysBetweenHuddles": 5,
        /* MaxDaysBetweenHuddles indicates the largest number of days that should pass between huddles in which this
           patient is discussed.  In other words, discuss this patient no less frequently than every 7 days. */
        "maxDaysBetweenHuddles": 7
      },
      /* This config indicates that patients with a risk score of 12-17 should be discussed about once every 3 weeks. */
      {
        "minScore": 12,
        "maxScore": 17,
        "minDaysBetweenHuddles": 15,
        "maxDaysBetweenHuddles": 21
      },
      /* This config indicates that patients with a risk score of 5-11 should be discussed about once every 6 weeks. */
      {
        "minScore": 5,
        "maxScore": 11,
        "minDaysBetweenHuddles": 36,
        "maxDaysBetweenHuddles": 42
      },
      /* This config indicates that patients with a risk score of 0-4 should be discussed about once every 13 weeks. */
      {
        "minScore": 0,
        "maxScore": 4,
        "minDaysBetweenHuddles": 85,
        "maxDaysBetweenHuddles": 91
      }
    ]
  },
  /* EventConfig contains the configuration for scheduling huddles based on events.  Currently only encounter events
     are supported.  This section can be removed if you do not wish to use events as a factor for scheduling huddles. */
  "eventConfig": {
    /* EncounterConfigs contains the configuration for scheduling huddles based on recent encounters.  The encounter's
       start date (admission) or end date (discharge) may be used. */
    "encounterConfigs": [
      {
        /* LookBackDays indicates how many days to look back when checking for recent encounters that might trigger a
           huddle discussion. */
        "lookBackDays": 7,
        /* TypeCodes indicates the encounter codes that should trigger a huddle discussion. */
        "typeCodes": [
          /* This config indicates that hospital discharges should trigger a huddle discussion. */
          {
            /* Name refers to a human friendly name for this scheduling rule.  It will be displayed as the reason for
               the patient being added to the huddle. */
            "name": "Hospital Discharge",
            /* System refers to the code system the encounter type is encoded in. */
            "system": "http://snomed.info/sct",
            /* Code refers to the specific code for the encounter type. */
            "code": "32485007",
            /* UseEndDate indicates if the end date of the encounter should be the triggering factor.  If useEndDate
               is not indicated or is false, the start date will be used instead. */
            "useEndDate": true
          },
          /* This config indicates that hospital admissions should trigger a huddle discussion. */
          {
            "name": "Hospital Admission",
            "system": "http://snomed.info/sct",
            "code": "32485007"
          },
          /* This config indicates that hospital re-admission discharges should trigger a huddle discussion. */
          {
            "name": "Hospital Re-Admission Discharge",
            "system": "http://snomed.info/sct",
            "code": "417005",
            "useEndDate": true
          },
          /* This config indicates that hospital re-admissions should trigger a huddle discussion. */
          {
            "name": "Hospital Re-Admission",
            "system": "http://snomed.info/sct",
            "code": "417005"
          },
          /* This config indicates that emergency room admissions should trigger a huddle discussion. */
          {
            "name": "Emergency Room Admission",
            "system": "http://snomed.info/sct",
            "code": "50849002"
          }
        ]
      }
    ]
  },
  /* SchedulerCronSpec indicates when the scheduling algorithm should be run.  The six digits corresond to seconds,
     minutes, hours, day of month, month, day of week.  In the example below, the algorithm is run every day at
     00:00:00.  For more information, see: https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format */
  "schedulerCronSpec": "0 0 0 * * *"
}

```